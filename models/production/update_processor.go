package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"github.com/theMomax/openefs/utils/metadata"

	"github.com/theMomax/openefs/config"
	"github.com/theMomax/openefs/models/production/weather"
	timeutils "github.com/theMomax/openefs/utils/time"
)

func init() {
	config.OnInitialize(func() {
		batchSize = config.Viper.GetUint(PathBatchSize)
		requiredSubsequent = batchSize - 1
		inferenceBatchSize = config.Viper.GetUint(PathInferenceBatchSize)
		requiredInferenceSubsequent = inferenceBatchSize - 1
		stepsize = config.Viper.GetDuration(PathStepSize)
		maxSize := batchSize
		if inferenceBatchSize > maxSize {
			maxSize = inferenceBatchSize
		}
		outdated = stepsize * time.Duration(maxSize)
		model = &metadata.Basic{
			Timestamp:  timeutils.Now(),
			Identifier: 0,
		}
		if _, err := os.Stat("./python/production.h5"); os.IsNotExist(err) {
			log.Info("creating production model...")
			cmd := exec.Command("python3", "./python/build_model_production.py", "./python/production.h5")
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.WithError(err).WithField("out", string(out)).Fatal("could not create production-model")
			}
			log.Debug("production-model created")
		}
	})
}

type cupdate struct {
	p Update
	w weather.Update
}

var (
	stepsize                    time.Duration
	batchSize                   uint
	requiredSubsequent          uint
	inferenceBatchSize          uint
	requiredInferenceSubsequent uint
	outdated                    time.Duration
)

var cache = make(map[time.Time]*cupdate)

var model metadata.Metadata

func handleProductionUpdate(u Update) {
	clearOutdatedCache()
	log.WithField("id", u.Meta().ID()).WithField("time", u.Time()).Debug("received production update")
	r := Round(u.Time())
	if cache[r] == nil {
		cache[r] = &cupdate{}
	}
	cache[r].p = u
	log.WithField("id", u.Meta().ID()).WithField("time", u.Time()).WithField("value", u.Data().Power).Trace("sending received update into outgoing channel")
	// send copy of actual data, so that changes are not reflected inside this file's logic
	outgoingProductionUpdates <- &update{
		time: u.Time(),
		meta: u.Meta(),
		data: &Data{
			Power:              u.Data().Power,
			nonNormalizedPower: u.Data().nonNormalizedPower,
		},
		derived: false,
	}
	log.Trace("\n" + formatCache(cache))
	applyUpdates()
}

func handleWeatherUpdate(wu weather.Update) {
	clearOutdatedCache()
	log.WithField("id", wu.Meta().ID()).WithField("time", wu.Time()).Trace("received weather update")
	r := Round(wu.Time())
	if cache[r] == nil {
		cache[r] = &cupdate{}
	}
	// do only apply update if it really contains new information
	if cache[r].w != nil && weather.Equal(cache[r].w.Data(), wu.Data()) {
		log.WithField("time", wu.Time()).Trace("dropped duplicate-update")
		return
	}
	cache[r].w = wu
	applyUpdates()
}

func applyUpdates() {
	timestamps := make([]time.Time, 0, len(cache))
	for t := range cache {
		timestamps = append(timestamps, t)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Sub(timestamps[j]) <= 0
	})
	log.WithField("cached_amount", len(timestamps)).Trace("applying updates...")

	modelDidChange := true

outer:
	for modelDidChange {
		modelDidChange = false
		log.Trace("checking for new update-possibilities since last iteration...")
		for i, t := range timestamps {
			c := cache[t]
			log.WithField("time", t).WithField("index", i).Trace("checking cached value: ", c)
			// does this step trigger a model-update?
			// the production-value does exist, and is newer than the model
			if batchSize != 0 && c != nil && c.p != nil && c.p.Meta().ID() > model.ID() {
				log.Trace("step triggers model-update")
				// can the model be updated?
				// both values exist for all required preceding and subsequent steps and this one, and there is no gap in the steps
				if forAllIs(timestamps, func(t time.Time) bool {
					return fullyExists(cache[t]) && !cache[t].p.IsDerived()
				}, rngI(i, i+int(requiredSubsequent))...) && isGapless(timestamps, i, i+int(requiredSubsequent), stepsize) {
					log.Trace("step fullfills requirements for model update")
					modelDidChange = training(t)
					continue outer
				}
			}

			// is this step to be predicted?
			// the value was not predicted yet, or ((the value was predicted with an older model or from older weather-data) and the value was not provided yet)
			if c == nil || c.p == nil || ((c.p.Meta().ID() < model.ID() || (c.w != nil && c.p.Meta().ID() < c.w.Meta().ID())) && c.p.IsDerived()) {
				log.WithField("time", t).Trace("step is to be predicted")

				// can this step be predicted?
				// the weather-value does exist for this timestamp, and both values exist for all required preceding steps, and there is no gap in the preceding steps
				if forAllIs(timestamps, func(t time.Time) bool {
					return weatherExists(cache[t]) && (c.p == nil || c.p.IsDerived())
				}, rngI(i, i+int(requiredInferenceSubsequent))...) && isGapless(timestamps, i, i+int(requiredInferenceSubsequent), stepsize) {
					log.Trace("step can be predicted")
					log.Trace(formatCache(cache))
					inference(t)
				}
			}
		}
	}
}

func inference(t time.Time) {
	log.WithField("time", t).Debug("starting inference...")
	args := []string{"./python/production.h5"}
	end := t.Add(time.Duration(requiredInferenceSubsequent-1) * stepsize)
	for i := t; end.Sub(i) >= 0; i = i.Add(stepsize) {
		args = append(args, formatTime(i)...)
		args = append(args, formatWeather(cache[i].w.Data())...)
	}

	log.Trace("calling python")
	cmd := exec.Command("python3", append([]string{"./python/inference_production.py"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).WithField("out", string(out)).WithField("cmd", cmd.String()).Error("inference on production model failed")
		return
	}

	var output [][][]float64
	elems := bytes.Split(bytes.TrimSpace(out), []byte("Model output:\n"))
	fout := []byte{}
	if len(elems) == 2 {
		fout = bytes.ReplaceAll(bytes.ReplaceAll(bytes.ReplaceAll(bytes.ReplaceAll(bytes.ReplaceAll(elems[1], []byte{'\n'}, []byte{','}), []byte{' '}, []byte{}), []byte{'0', '.', ','}, []byte{'0', '.', '0', ','}), []byte{'0', '.', ']'}, []byte{'0', '.', '0', ']'}), []byte{',', ','}, []byte{','})
		err = json.Unmarshal(fout, &output)
	} else {
		err = errors.New("inference-ouput does not contain result-section")
	}
	if err != nil {
		if len(elems) == 2 {
			log.WithError(err).WithField("out", string(fout)).Error("inference on production model failed")
		} else {
			log.WithError(err).WithField("out", string(out)).Error("inference on production model failed")
		}
		return
	}

	log.WithField("output", output).Trace("call to python completed")

	if cache[t] == nil {
		cache[t] = &cupdate{}
	}

	for i := range output {
		t := t.Add(time.Duration(i) * stepsize)
		m := latest(model, cache[t].w.Meta())
		cache[t].p = &update{
			data: &Data{
				Power: output[i][0][0],
			},
			time:    t,
			meta:    m,
			derived: true,
		}
		log.WithField("id", cache[t].p.Meta().ID()).WithField("time", cache[t].p.Time()).WithField("value", cache[t].p.Data().Power).WithField("id", cache[t].p.Meta().ID()).Trace("sending update into outgoing channel")
		outgoingProductionUpdates <- cache[t].p
	}

	log.Debug("predicted production-values")
}

func training(t time.Time) (ok bool) {
	log.WithField("time", t).Debug("starting training...")
	latest := model

	args := []string{"./python/production.h5"}
	end := t.Add(time.Duration(requiredSubsequent) * stepsize)
	for i := t; end.Sub(i) >= 0; i = i.Add(stepsize) {
		if cache[i].w.Meta().ID() > latest.ID() {
			latest = cache[i].w.Meta()
		}
		if cache[i].p.Meta().ID() > latest.ID() {
			latest = cache[i].p.Meta()
		}
		args = append(args, formatTime(i)...)
		args = append(args, formatWeather(cache[i].w.Data())...)
		args = append(args, formatProduction(cache[i].p.Data())...)
		fmt.Println("target:", formatProduction(cache[i].p.Data()), " (", i.Hour(), ")")
	}

	log.Trace("calling python")
	cmd := exec.Command("python3", append([]string{"./python/training_production.py"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.WithError(err).WithField("out", string(out)).WithField("cmd", cmd.String()).Error("training on production model failed")
		return false
	}
	model = latest
	log.WithField("model", model).Trace("model updated")
	log.WithField("id", model.ID()).WithField("time", t).Debug("updated production-model")
	return true
}

func formatProduction(p *Data) []string {
	if p == nil {
		return []string{}
	}
	return []string{ff(p.Power)}
}

func formatWeather(w *weather.Data) []string {
	if w == nil {
		return []string{}
	}
	return []string{ff(w.CloudCover), ff(w.PrecipitationProbability), ff(w.WindSpeed), ff(w.WindGust), ff(w.PrecipitationIntensity), ff(w.ApparentTemperature), ff(w.Humidity), ff(w.DewPoint), ff(w.Visibility), ff(w.UVIndex), ff(w.Temperature)}
}

func formatTime(t time.Time) []string {
	return []string{ff(timeutils.YearProcess(t)), ff(timeutils.DayProcess(t))}
}

func ff(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

func clearOutdatedCache() {
	before := len(cache)
	for t := range cache {
		if timeutils.Now().Sub(t) >= outdated {
			delete(cache, t)
		}
	}
	if before > len(cache) {
		log.WithField("before", before).WithField("after", len(cache)).WithField("deleted", before-len(cache)).Debug("cleared outdated cache")
	}
}

func fullyExists(c *cupdate) bool {
	return c != nil && c.p != nil && c.w != nil
}

func weatherExists(c *cupdate) bool {
	return c != nil && c.w != nil
}

func rng(from, to int) []int {
	if from >= to {
		return []int{}
	}
	numbers := make([]int, 0, to-from)
	for i := from; i < to; i++ {
		numbers = append(numbers, i)
	}
	return numbers
}

func rngI(from, toInclusive int) []int {
	return rng(from, toInclusive+1)
}

func isGapless(timeline []time.Time, start, end int, unit time.Duration) bool {
	if start > end || start < 0 || end >= len(timeline) {
		return false
	}
	if start == end {
		return true
	}

	return timeline[end].Sub(timeline[start]) == time.Duration(end-start)*unit
}

func forAllIs(timeline []time.Time, condition func(time.Time) bool, indices ...int) bool {
	for _, i := range indices {
		if i < 0 || i >= len(timeline) || !condition(timeline[i]) {
			return false
		}
	}
	return true
}

func formatCache(cache map[time.Time]*cupdate) string {
	s := "time\t\t\t\t | ahead\t\t\t\t | production\t\t\t\t | weather\t\t\t\t | prodDerived\n========================================================================================================================================\n"

	timestamps := make([]time.Time, 0, len(cache))
	for t := range cache {
		timestamps = append(timestamps, t)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Sub(timestamps[j]) <= 0
	})

	for i, t := range timestamps {
		if !isGapless(timestamps, i-1, i, stepsize) {
			s += "GAP\n"
		}
		s += t.Format(time.ANSIC) + "\t | "
		s += t.Sub(timeutils.Now()).String() + "\t | "
		c := cache[t]
		if c == nil {
			s += "nil\n"
		} else {
			if c.p == nil {
				s += "nil\t\t\t\t | "
			} else {
				s += ff(c.p.Data().Power) + " (" + strconv.Itoa(int(c.p.Meta().ID())) + ")" + "\t\t\t\t | "
			}
			if c.w == nil {
				s += "nil\t\t\t\t"
			} else {
				// b, _ := json.Marshal(c.w.Data())

				s += "some" + " (" + strconv.Itoa(int(c.w.Meta().ID())) + ")" + "\t\t\t\t"
			}
			if c.p != nil && c.p.IsDerived() {
				s += "true\n"
			} else {
				s += "false\n"
			}
		}
	}
	return s
}

func latest(first metadata.Metadata, others ...metadata.Metadata) metadata.Metadata {
	latest := first
	for _, o := range others {
		if latest.ID() < o.ID() {
			latest = o
		}
	}
	return latest
}
