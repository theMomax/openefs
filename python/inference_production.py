#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np
import json

INPUT_SHAPE = (None, 13)
OUTPUT_SHAPE = 1

if (((len(sys.argv) - 2) % (INPUT_SHAPE[1]) != 0) or ((len(sys.argv) - 2) == 0)) :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be a multiple of INPUT_SHAPE[1]: ' + str(INPUT_SHAPE[1]) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)


batch_input = []

i = 2
while (i <= len(sys.argv) - (INPUT_SHAPE[1])):
    features = []
    for j in range(0, INPUT_SHAPE[1]):
        features.append(float(sys.argv[i+j]))
    
    batch_input.append([features])
    i += INPUT_SHAPE[1]



model = tf.keras.models.load_model(sys.argv[1])

model_input = np.asarray(batch_input)
print('Model input:')
print(model_input)

model_output = model.predict(model_input)

print('Model output:')
print(model_output)