#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np
import json

INPUT_SHAPE = (2, 14)
OUTPUT_SHAPE = 1

if (((len(sys.argv) - 2 - (INPUT_SHAPE[0] * OUTPUT_SHAPE)) % (INPUT_SHAPE[1]-1) != 0) and ((len(sys.argv) - 2 - (INPUT_SHAPE[0] * OUTPUT_SHAPE)) != 0)) :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be INPUT_SHAPE[0] + OUTPUT_SHAPE (' + str(INPUT_SHAPE[0] * OUTPUT_SHAPE) + ') plus a multiple of INPUT_SHAPE[1]-OUTPUT_SHAPE: ' + str(INPUT_SHAPE[1]-OUTPUT_SHAPE) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)


def time(i):
    time = []
    for j in range(0, 2):
        time.append(float(sys.argv[2+(INPUT_SHAPE[0]*OUTPUT_SHAPE)+(i*(INPUT_SHAPE[1]-1))+j]))
    
    return time

def production(i):
    return float(sys.argv[2+i])

def weather(i):
    weather = []
    for j in range(2, INPUT_SHAPE[1]-1):
        weather.append(float(sys.argv[2+(INPUT_SHAPE[0]*OUTPUT_SHAPE)+(i*(INPUT_SHAPE[1]-1))+j]))
    
    return weather



# setup initial input_data
input_data = []

for i in range(0, INPUT_SHAPE[0]):
    # time-data(2) + production(1) + weather(11)
    features = []
    for t in time(i):
        features.append(t)

    features.append(production(i))

    for w in weather(i):
        features.append(w)

    input_data.append(features)


iterations = int((len(sys.argv) - 2 - (INPUT_SHAPE[0] * OUTPUT_SHAPE)) / (INPUT_SHAPE[1]-OUTPUT_SHAPE) - 1)

model = tf.keras.models.load_model(sys.argv[1])
model_output = []

for i in range(2, iterations+2):
    model_input = np.asarray([input_data])
    print('Model input:')
    print(model_input)


    out = model.predict(model_input)[0][0]
    print('Out:')
    print(out)
    model_output.append(out)

    if (i < iterations+1):
        for j in range(0, INPUT_SHAPE[0]-1):
            input_data[j] = input_data[j+1]

        features = []
        for t in time(i):
            features.append(t)

        features.append(out)

        for w in weather(i):
            features.append(w)

        input_data[len(input_data)-1] = features



print('Model output:')
print(model_output)



# batch = []

# i = 2
# while (i <= len(sys.argv) - (INPUT_SHAPE[0] * INPUT_SHAPE[1])):
#     timesteps = []
#     j = 0
#     while (j < INPUT_SHAPE[1] * INPUT_SHAPE[0]):
#         features = []
#         for k in range(0,INPUT_SHAPE[1]):
#             features.append(float(sys.argv[i+j+k]))

#         j += INPUT_SHAPE[1]
#         timesteps.append(features)

#     i += (INPUT_SHAPE[0] * INPUT_SHAPE[1])
#     batch.append(timesteps)