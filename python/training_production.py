#!/usr/bin/python

import sys
import tensorflow as tf
import numpy as np
import tensorflow.keras.backend as K

INPUT_SHAPE = (None,13)
OUTPUT_SHAPE = 1

if (len(sys.argv) - 2) % (INPUT_SHAPE[1] + OUTPUT_SHAPE) != 0 or ((len(sys.argv) - 2) == 0) :
    print('Illegal number of arguments: expected <ModelPath> <InputValues>... (Number must be a multiple of INPUT_SHAPE[1] + OUTPUT_SHAPE: ' + str(INPUT_SHAPE[1] + OUTPUT_SHAPE) + ')' + ' got ' + str(len(sys.argv)-2))
    exit(1)

batch_input = []
batch_target = []

i = 2
while (i <= len(sys.argv) - (INPUT_SHAPE[1] + OUTPUT_SHAPE)):
    features = []
    for j in range(0, INPUT_SHAPE[1]):
        features.append(float(sys.argv[i+j]))
    
    batch_input.append([features])
    batch_target.append(float(sys.argv[i+INPUT_SHAPE[1]]))
    i += INPUT_SHAPE[1]+OUTPUT_SHAPE


model_input = np.asarray(batch_input)
print('Model input:')
print(model_input)

model_target = np.asarray(batch_target)
print('Model target:')
print(model_target)

model = tf.keras.models.load_model(sys.argv[1])
# K.set_value(model.optimizer.lr, 0.001)
model.fit(model_input, model_target, 
    epochs=200,
    steps_per_epoch=1,
    shuffle=True,
)

model.save(sys.argv[1])

print('Saved production-model to ' + sys.argv[1] + ' !')