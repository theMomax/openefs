#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np

INPUT_SHAPE = (3,13)
OUTPUT_SHAPE = 1

if (len(sys.argv) - 2) % (INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE) != 0 :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be a multiple of INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE: ' + str(INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)

batch_input = []
batch_target = []

i = 2
while (i <= len(sys.argv) - (INPUT_SHAPE[0] * INPUT_SHAPE[1]) + OUTPUT_SHAPE):
    timesteps = []
    j = 0
    while (j < INPUT_SHAPE[1] * INPUT_SHAPE[0]):
        features = []
        for k in range(0,INPUT_SHAPE[1]):
            features.append(float(sys.argv[i+j+k]))

        j += INPUT_SHAPE[1]
        timesteps.append(features)

    i += (INPUT_SHAPE[0] * INPUT_SHAPE[1] + OUTPUT_SHAPE)
    batch_input.append(timesteps)
    batch_target.append(float(sys.argv[i-1]))



model_input = np.asarray(batch_input)
print('Model input:')
print(model_input)

model_target = np.asarray(batch_target)
print('Model target:')
print(model_target)

model = tf.keras.models.load_model(sys.argv[1])
model.fit(model_input, model_target, 
    epochs=40,
    steps_per_epoch=1,
    shuffle=False,
)

model.save(sys.argv[1])

print('Saved production-model to ' + sys.argv[1] + ' !')