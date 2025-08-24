import tensorflow as tf
from tensorflow import keras

# Загружаем датасет MNIST
(x_train, y_train), (x_test, y_test) = keras.datasets.mnist.load_data()

# Нормализация
x_train = x_train.astype("float32") / 255.0
x_test = x_test.astype("float32") / 255.0

# Добавляем канал (28,28,1)
x_train = x_train[..., tf.newaxis]
x_test = x_test[..., tf.newaxis]

# Архитектура модели
model = keras.Sequential([
    keras.layers.Conv2D(32, (3,3), activation="relu", input_shape=(28,28,1)),
    keras.layers.MaxPooling2D((2,2)),
    keras.layers.Conv2D(64, (3,3), activation="relu"),
    keras.layers.MaxPooling2D((2,2)),
    keras.layers.Flatten(),
    keras.layers.Dense(128, activation="relu"),
    keras.layers.Dropout(0.5),
    keras.layers.Dense(10, activation="softmax")
])

# Компиляция
model.compile(optimizer="adam",
              loss="sparse_categorical_crossentropy",
              metrics=["accuracy"])

# Обучение
model.fit(x_train, y_train, epochs=5, validation_split=0.1)

# Оценка
test_loss, test_acc = model.evaluate(x_test, y_test, verbose=2)
print(f"✅ Test accuracy: {test_acc:.4f}")

# Сохраняем модель
model.save("models/my_model.keras")
