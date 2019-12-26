#!/usr/bin/python

import sys
import tensorflow as tf


INPUT_SHAPE = (13,1)
OUTPUT_SHAPE = 1

if len(sys.argv) != 2:
    print('Illegal number of arguments: expected <OutputPath>')
    exit(1)


model = tf.keras.models.Sequential()
model.add(tf.keras.layers.LSTM(32,input_shape=INPUT_SHAPE))
model.add(tf.keras.layers.Dense(OUTPUT_SHAPE))

model.compile(optimizer=tf.keras.optimizers.RMSprop(), loss='mae')

model.save(sys.argv[1])

print('Saved production-model to ' + sys.argv[1] + ' !')