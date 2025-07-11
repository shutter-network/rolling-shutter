from itertools import combinations

def allcombinations(size):
    for i in range(size + 1):
        for c in combinations(range(size), i):
            yield c


def allcombinations_upto(maxsize):
    for i in range(maxsize + 1):
        for c in allcombinations(i):
            yield (i, c)


for i, (log, indexed) in enumerate(allcombinations_upto(4)):
    topics = " ".join(f"topic{j}" for j in indexed)
    print(f"{i:02x} LOG{log} {topics}")

# opcode-matching-topics-table
# i.e. '0b' means the event has LOG3 opcode and we want to match topic0 and topic1 (but not topic2)
"""
00 LOG0
01 LOG1
02 LOG1 topic0
03 LOG2
04 LOG2 topic0
05 LOG2 topic1
06 LOG2 topic0 topic1
07 LOG3
08 LOG3 topic0
09 LOG3 topic1
0a LOG3 topic2
0b LOG3 topic0 topic1
0c LOG3 topic0 topic2
0d LOG3 topic1 topic2
0e LOG3 topic0 topic1 topic2
0f LOG4
10 LOG4 topic0
11 LOG4 topic1
12 LOG4 topic2
13 LOG4 topic3
14 LOG4 topic0 topic1
15 LOG4 topic0 topic2
16 LOG4 topic0 topic3
17 LOG4 topic1 topic2
18 LOG4 topic1 topic3
19 LOG4 topic2 topic3
1a LOG4 topic0 topic1 topic2
1b LOG4 topic0 topic1 topic3
1c LOG4 topic0 topic2 topic3
1d LOG4 topic1 topic2 topic3
1e LOG4 topic0 topic1 topic2 topic3
"""
