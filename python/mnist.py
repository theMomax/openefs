from __future__ import absolute_import, division, print_function, unicode_literals

# Install TensorFlow

import tensorflow as tf

mnist = tf.keras.datasets.mnist

(x_train, y_train), (x_test, y_test) = mnist.load_data()
x_train, x_test = x_train / 255.0, x_test / 255.0

sess = tf.compat.v1.Session()
tf.compat.v1.keras.backend.set_session(sess)

model = tf.keras.models.Sequential([
  tf.keras.layers.Flatten(input_shape=(28, 28), name="inputLayer"),
  tf.keras.layers.Dense(128, activation='relu'),
  tf.keras.layers.Dropout(0.2),
  tf.keras.layers.Dense(10, activation='softmax', name="inferenceLayer")
])

model.compile(optimizer='adam',
              loss='sparse_categorical_crossentropy',
              metrics=['accuracy'])

model.fit(x_train, y_train, epochs=5)

model.evaluate(x_test,  y_test, verbose=2)

builder = tf.compat.v1.saved_model.builder.SavedModelBuilder("mnistmodel")
builder.add_meta_graph_and_variables(sess, ["mnist"])
builder.save()
sess.close()