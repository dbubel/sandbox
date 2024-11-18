import pandas as pd
import msgpack

# Assuming you have a DataFrame called 'df'
df = pd.DataFrame({"A": [1, 2, 3], "B": [4, 5, 6]})

# Convert DataFrame to a dict
data = df.to_dict(orient="split")

# Serialize to MessagePack
packed_data = msgpack.packb(data)

# Save to a file
with open("data.msgpack", "wb") as f:
    f.write(packed_data)
