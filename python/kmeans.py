import json
from sklearn.cluster import KMeans
import numpy as np
import time


def load_vectors_from_jsonl(file_path):
    vectors = []
    with open(file_path, 'r') as file:
        for line in file:
            vector = json.loads(line)
            vectors.append(vector)
    return np.array(vectors)


def perform_kmeans_clustering(vectors, num_clusters, epsilon=1e-1):
    kmeans = KMeans(n_clusters=num_clusters, random_state=42, tol=epsilon)
    kmeans.fit(vectors)
    return kmeans.labels_, kmeans.cluster_centers_


def main():
    input_file = '../data/8_f32_rand_1m.jsonl'  # Path to your JSON Lines file
    num_clusters = 10  # Number of clusters to form

    start_time = time.time()
    vectors = load_vectors_from_jsonl(input_file)
    end_time = time.time()
    elapsed_time = end_time - start_time
    print(f"\nTime read: {elapsed_time:.4f} seconds")

    start_time = time.time()
    labels, cluster_centers = perform_kmeans_clustering(vectors, num_clusters)
    end_time = time.time()
    elapsed_time = (end_time - start_time) * 1000  # Convert to milliseconds
    print(f"\ncluster time: {elapsed_time:.4f} ms")
    print("Cluster Labels for Each Vector:")
    print(labels)
    print("\nCluster Centers:")
    print(cluster_centers)


if __name__ == "__main__":
    main()
