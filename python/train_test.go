package python_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/magiconair/properties/assert"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// from https://tonytruong.net/running-a-keras-tensorflow-model-in-golang/
func TestTensorflowIntegrationMNIST(t *testing.T) {
	// replace myModel and myTag with the appropriate exported names in the chestrays-keras-binary-classification.ipynb
	model, err := tf.LoadSavedModel("mnistmodel", []string{"mnist"}, nil)

	if err != nil {
		fmt.Printf("Error loading saved model: %s\n", err.Error())
		t.Fail()
		return
	}

	defer model.Session.Close()

	input, _ := tf.NewTensor([1][28][28][1]float32{})

	result, err := model.Session.Run(
		map[tf.Output]*tf.Tensor{
			model.Graph.Operation("inputLayer_input").Output(0): input, // Replace this with your input layer name
		},
		[]tf.Output{
			model.Graph.Operation("inferenceLayer/Softmax").Output(0), // Replace this with your output layer name
		},
		nil,
	)

	if err != nil {
		fmt.Printf("Error running the session with input, err: %s\n", err.Error())
		t.Fail()
		return
	}

	switch reflect.TypeOf(result[0].Value()).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(result[0].Value())
		assert.Equal(t, 10, s.Index(0).Len())
	default:
		fmt.Println("result-value is no slice")
		t.Fail()
	}
}
