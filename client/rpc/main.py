# import grpc
# import time
# from concurrent import futures
# import Content_pb2 as pbContent
# import Content_pb2_grpc as grpcContent
# from sentence_transformers import SentenceTransformer
# from pymilvus import (
#     CollectionSchema,
#     MilvusClient,
#     FieldSchema,
#     DataType,
#     Collection,
#     connections,
# )
#
# # COUNT = 1000
# BATCH_SIZE = 64
# TOP_K = 100
# COLLECTION_NAME = "content"
# EXTERNAL_SERVER_PORT = 50051
# INTERNAL_SERVER_PORT = 50052
# _ONE_DAY_IN_SECONDS = 60 * 60 * 24
#
# # client = MilvusClient("http://localhost:19530")
# connections.connect(host="localhost", port=19530)
# schema = CollectionSchema(
#     description="content search",
#     fields=[
#         FieldSchema(
#             "content_id",
#             DataType.VARCHAR,
#             is_primary=True,
#             auto_id=False,
#             max_length=64,
#         ),
#         FieldSchema(
#             "text",
#             DataType.VARCHAR,
#             max_length=2**16 - 1,
#         ),
#         FieldSchema(
#             "embedding",
#             DataType.FLOAT_VECTOR,
#             dim=384,
#         ),
#     ],
# )
# collection = Collection(name=COLLECTION_NAME, schema=schema)
# index_params = {
#     "metric_type": "L2",
#     "index_type": "IVF_FLAT",
#     "params": {"nlist": 1536},
# }
# collection.create_index(field_name="embedding", index_params=index_params)
# collection.load()
#
#
# transformer = SentenceTransformer("all-MiniLM-L6-v2")
#
#
# def embed_insert(data):
#     content_ids = data[0]
#     texts = data[1]
#     embeds = transformer.encode(data[1])
#     embeddings = [x for x in embeds]
#
#     ins = [content_ids, texts, embeddings]
#     res = collection.insert(ins)
#     print(res)
#
#
# def embed_search(data):
#     embeds = transformer.encode(data)
#     return [x for x in embeds]
#
#
# class ContentService:
#     def Pop(self, request, context):
#         context.set_code(grpc.StatusCode.UNIMPLEMENTED)
#         context.set_details("Method not implemented!")
#         return NotImplementedError("Method not implemented!")
#
#     def Rollback(self, request, context):
#         res = collection.delete(f"content_id in {request.items}")
#         print(res)
#         print("total content:", collection.num_entities)
#
#     def InsertContents(self, request, context):
#         data_batch = [[], []]
#
#         try:
#             for i, _ in enumerate(request.ids):
#                 id = request.ids[i]
#                 text = request.texts[i].strip()
#                 data_batch[0].append(id)
#                 data_batch[1].append(text)
#                 if len(data_batch[0]) % BATCH_SIZE == 0:
#                     embed_insert(data_batch)
#                     data_batch = [[], []]
#
#             if len(data_batch) != 0 and len(data_batch[0]) != 0:
#                 embed_insert(data_batch)
#             # collection.flush()
#             print("total content:", collection.num_entities)
#         except Exception as e:
#             print(data_batch)
#             raise e
#
#         status = pbContent.Status()
#         status.code = 0
#         return status
#
#     def SemanticSearch(self, request, context):
#         search_data = [embed_search(request.text)]
#         output = collection.search(
#             data=search_data,
#             anns_field="embedding",
#             param={},
#             limit=TOP_K,
#             output_fields=["content_id", "text"],
#         )
#
#         result = pbContent.ContentIDs()
#         for out in output:
#             for o in out:
#                 content_id = o.entity.get("content_id")
#                 result.items.append(content_id)
#
#         return result
#
#
# def server():
#     server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
#     grpcContent.add_ContentServicer_to_server(ContentService(), server)
#     server.add_insecure_port(f"[::]:{INTERNAL_SERVER_PORT}")
#     server.start()
#     print("Server started. Listen on port:", INTERNAL_SERVER_PORT)
#     # server.wait_for_termination()
#     try:
#         while True:
#             time.sleep(_ONE_DAY_IN_SECONDS)
#     except KeyboardInterrupt:
#         server.stop(0)
#
#
# def cl():
#     # Open a gRPC channel
#     channel = grpc.insecure_channel(f"localhost:{EXTERNAL_SERVER_PORT}")
#
#     # Create a stub for the Content service
#     stub = grpcContent.ContentStub(channel)
#
#     # Make a request to the Pop RPC method
#     persona_id = pbContent.Empty()
#     content_ids = stub.Pop(persona_id)
#
#     # Print the response
#     print("Content IDs:", content_ids.items)
#
#
# def main():
#     server()
#
#
# if __name__ == "__main__":
#     main()
