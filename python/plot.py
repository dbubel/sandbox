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
    for item in data:
        point = item["data"]
        color = item["color"]

        # Plotting the point
        plt.scatter(point[0], point[1], color=color, label=f'Point ({color})')

    # Remove duplicate labels in the legend
    handles, labels = plt.gca().get_legend_handles_labels()
    by_label = dict(zip(labels, handles))
    plt.legend(by_label.values(), by_label.keys())

    plt.xlabel('X-axis')
    plt.ylabel('Y-axis')
    plt.title('Plot of Points')
    plt.show()


if __name__ == "__main__":
    data = read_jsonl_from_stdin()

    # Sample the data if necessary
    sample_size = 100  # Set the desired sample size here
    data = sample_data(data, sample_size)

    plot_data(data)
