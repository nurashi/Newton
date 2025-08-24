import grpc
import proto.ml_pb2 as ml_service_pb2
import proto.ml_pb2_grpc as ml_service_pb2_grpc

def run():
    with grpc.insecure_channel("localhost:50051") as channel:
        stub = ml_service_pb2_grpc.MLServiceStub(channel)

        with open("pixil-frame-0.png", "rb") as f:
            img_bytes = f.read()

        response = stub.Predict(ml_service_pb2.PredictRequest(image=img_bytes))
        print(f"Predicted digit: {response.digit}")

if __name__ == "__main__":
    run()
