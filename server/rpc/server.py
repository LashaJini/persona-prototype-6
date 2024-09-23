from transformers import pipeline
import grpc
import time
from concurrent import futures
import Content_pb2 as pbContent
import Content_pb2_grpc as grpcContent
from sentence_transformers import SentenceTransformer
from pymilvus import (
    CollectionSchema,
    FieldSchema,
    DataType,
    Collection,
    connections,
)

# COUNT = 1000
BATCH_SIZE = 64
TOP_K = 100
COLLECTION_NAME = "content"
PYTHON_GRPC_PORT = 50052
SPAM_LABEL = "LABEL_1"
_ONE_DAY_IN_SECONDS = 60 * 60 * 24

# model can't handle more
max_spam_sentence_length = 512
spam_pipeline = pipeline(
    "text-classification", model="Titeiiko/OTIS-Official-Spam-Model"
)
# emotions_pipeline = pipeline("text-classification", model="SamLowe/roberta-base-go_emotions", top_k=1)
# personalities_pipeline = pipeline("text-classification", model="Minej/bert-base-personality", top_k=5)

# client = MilvusClient("http://localhost:19530")
connections.connect(host="localhost", port=19530)
schema = CollectionSchema(
    description="content search",
    fields=[
        FieldSchema(
            "content_id",
            DataType.VARCHAR,
            is_primary=True,
            auto_id=False,
            max_length=64,
        ),
        FieldSchema(
            "text",
            DataType.VARCHAR,
            max_length=2**16 - 1,
        ),
        FieldSchema(
            "embedding",
            DataType.FLOAT_VECTOR,
            dim=384,
        ),
    ],
)
collection = Collection(name=COLLECTION_NAME, schema=schema)
index_params = {
    "metric_type": "L2",
    "index_type": "IVF_FLAT",
    "params": {"nlist": 1536},
}
collection.create_index(field_name="embedding", index_params=index_params)
collection.load()


transformer = SentenceTransformer("all-MiniLM-L6-v2")


def embed_insert(data):
    content_ids = data[0]
    texts = data[1]
    embeds = transformer.encode(data[1])
    embeddings = [x for x in embeds]

    ins = [content_ids, texts, embeddings]
    res = collection.insert(ins)
    print(res)


def embed_search(data):
    embeds = transformer.encode(data)
    return [x for x in embeds]


class ContentService:
    def Personalize(self, request, context):
        truncated_text = [text[:max_spam_sentence_length] for text in request.texts]
        spam_outputs = spam_pipeline(truncated_text)
        # emotions_outputs = emotions_pipeline(truncated_text)
        # personalities_outputs = personalities_pipeline(truncated_text)

        if spam_outputs:
            for output in spam_outputs:
                print(output)

        # if emotions_outputs:
        #     for output in emotions_outputs:
        #         print(output)

        # if personalities_outputs:
        #     for output in personalities_outputs:
        #         print(output)


    def CalculateSpamProbs(self, request, context):
        sentences = request.items
        truncated_sentences = [
            sentence[:max_spam_sentence_length] for sentence in sentences
        ]

        probs = pbContent.SpamProbs()
        if len(truncated_sentences) > 0:
            spam_outputs = spam_pipeline(truncated_sentences)

            if spam_outputs:
                for index, output in enumerate(spam_outputs):
                    label = output["label"]
                    score = output["score"]

                    # INFO: empty POST does not mean a spam in Reddit's case
                    if len(truncated_sentences[index]) == 0:
                        probs.items.append(0)
                        continue

                    if label == SPAM_LABEL:
                        probs.items.append(float(f"{score:.4f}"))
                    else:
                        probs.items.append(float(f"{1-score:.4f}"))
        return probs

    def Rollback(self, request, context):
        res = collection.delete(f"content_id in {request.items}")
        print(res)
        print("total content:", collection.num_entities)

    def InsertContents(self, request, context):
        data_batch = [[], []]

        try:
            for i, _ in enumerate(request.ids):
                id = request.ids[i]
                text = request.texts[i].strip()
                data_batch[0].append(id)
                data_batch[1].append(text)
                if len(data_batch[0]) % BATCH_SIZE == 0:
                    embed_insert(data_batch)
                    data_batch = [[], []]

            if len(data_batch) != 0 and len(data_batch[0]) != 0:
                embed_insert(data_batch)
            # collection.flush()
            print("total content:", collection.num_entities)
        except Exception as e:
            print(data_batch)
            raise e

        status = pbContent.Status()
        status.code = 0
        return status

    def SemanticSearch(self, request, context):
        search_data = [embed_search(request.text)]
        output = collection.search(
            data=search_data,
            anns_field="embedding",
            param={},
            limit=TOP_K,
            output_fields=["content_id", "text"],
        )

        result = pbContent.ContentIDs()
        for out in output:
            for o in out:
                content_id = o.entity.get("content_id")
                result.items.append(content_id)

        return result


def server():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    grpcContent.add_ContentServicer_to_server(ContentService(), server)
    server.add_insecure_port(f"[::]:{PYTHON_GRPC_PORT}")
    server.start()
    print("Server started. Listen on port:", PYTHON_GRPC_PORT)
    # server.wait_for_termination()
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)


def main():
    server()


if __name__ == "__main__":
    main()
