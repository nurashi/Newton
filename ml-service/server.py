import grpc
from concurrent import futures
import numpy as np
import tensorflow as tf
from tensorflow import keras
import proto.ml_pb2 as ml_service_pb2
import proto.ml_pb2_grpc as ml_service_pb2_grpc

model = keras.models.load_model("models/my_model.keras")

class MLService(ml_service_pb2_grpc.MLServiceServicer):
    def Predict(self, request, context):
        img = tf.io.decode_image(request.image, channels=1)
        img = tf.image.resize(img, [28, 28])
        img = img.numpy().reshape(1, 28, 28) / 255.0

        prediction = np.argmax(model.predict(img))
        return ml_service_pb2.PredictResponse(digit=int(prediction))

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    ml_service_pb2_grpc.add_MLServiceServicer_to_server(MLService(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    print("🚀 ML gRPC server running on port 50051")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
