import sys
import json
import random
import matplotlib.pyplot as plt


def read_jsonl_from_stdin():
    data = []
    for line in sys.stdin:
        data.append(json.loads(line))
    return data


def sample_data(data, sample_size=None):
    if sample_size is None or sample_size >= len(data):
        return data
    return random.sample(data, sample_size)


def plot_data(data):
    points = {}
    centroids = {}

    # Extract points and centroids
    for item in data:
        vec = item["vec"]
        centroid = tuple(item["centroid"])
        if centroid not in points:
            points[centroid] = []

        points[centroid].append(vec)
        centroids[centroid] = centroid

    # Define a list of colors for different groups
    colors = ["blue", "green", "orange", "purple", "cyan",
              "magenta", "yellow", "brown", "pink", "gray"]

    # Plot points with different colors for each group
    for idx, (centroid, group_points) in enumerate(points.items()):
        color = colors[idx % len(colors)]
        for point in group_points:
            plt.scatter(point[0], point[1], color=color, label=f'Group {idx}' if idx not in plt.gca(
            ).get_legend_handles_labels()[1] else "")

    # Plot centroids in black with a different marker shape
    for centroid in centroids.values():
        plt.scatter(centroid[0], centroid[1], color='black', s=100, marker='X',
                    label='Centroid' if 'Centroid' not in plt.gca().get_legend_handles_labels()[1] else "")

    # Remove duplicate labels in the legend
    handles, labels = plt.gca().get_legend_handles_labels()
    by_label = dict(zip(labels, handles))
    plt.legend(by_label.values(), by_label.keys())

    plt.xlabel('X-axis')
    plt.ylabel('Y-axis')
    plt.title('Plot of K-Means Clustering Results')
    plt.show()


if __name__ == "__main__":
    data = read_jsonl_from_stdin()

    # Sample the data if necessary
    sample_size = 1000  # Set the desired sample size here
    data = sample_data(data, sample_size)

    plot_data(data)
