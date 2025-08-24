import tensorflow as tf
import numpy as np

model = tf.keras.models.load_model("model.h5")

(x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
x_test = x_test.astype("float32") / 255.0
x_test = np.expand_dims(x_test, -1)

pred = model.predict(x_test[:5])
print("Predictions:", np.argmax(pred, axis=1))
print("Labels:     ", y_test[:5])
