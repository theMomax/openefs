#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np

INPUT_SHAPE = (13,1)
OUTPUT_SHAPE = 1

if (len(sys.argv) - 2) % (INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE) != 0 :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be a multiple of INPUT_SHAPE + OUTPUT_SHAPE: ' + str(INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)

timestamps_input = []
timestamps_target = []


i = 2
while (i <= len(sys.argv) - (INPUT_SHAPE[0] * INPUT_SHAPE[1]) + OUTPUT_SHAPE):
    features = []
    j = 0
    while (j < INPUT_SHAPE[1] * INPUT_SHAPE[0]):
        values = []
        for k in range(0,INPUT_SHAPE[1]):
            values.append(float(sys.argv[i+j+k]))

        j += INPUT_SHAPE[1]
        features.append(values)

    i += (INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE)
    timestamps_input.append(features)
    timestamps_target.append(float(sys.argv[i-1]))

model_input = np.asarray(timestamps_input)
print('Model input:')
print(model_input)

model_target = np.asarray(timestamps_target)
print('Model target:')
print(model_target)

model = tf.keras.models.load_model(sys.argv[1])
model.fit(model_input, model_target, 
    epochs=50,
    steps_per_epoch=1,
)

model.save(sys.argv[1])

print('Saved production-model to ' + sys.argv[1] + ' !')