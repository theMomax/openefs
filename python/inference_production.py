#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np

INPUT_SHAPE = (13,1)
OUTPUT_SHAPE = 1

if (len(sys.argv) - 2) % (INPUT_SHAPE[0] * INPUT_SHAPE[1]) != 0 :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be a multiple of INPUT_SHAPE: ' + str(INPUT_SHAPE[0] * INPUT_SHAPE[1]) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)

timestamps = []

i = 2
while (i <= len(sys.argv) - (INPUT_SHAPE[0] * INPUT_SHAPE[1])):
    features = []
    j = 0
    while (j < INPUT_SHAPE[1] * INPUT_SHAPE[0]):
        values = []
        for k in range(0,INPUT_SHAPE[1]):
            values.append(float(sys.argv[i+j+k]))

        j += INPUT_SHAPE[1]
        features.append(values)

    i += (INPUT_SHAPE[0] * INPUT_SHAPE[1])
    timestamps.append(features)

model_input = np.asarray(timestamps)
print('Model input:')
print(model_input)

model = tf.keras.models.load_model(sys.argv[1])

print('Model output:')
print(model.predict(model_input))