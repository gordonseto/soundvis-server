import pandas as pd
from surprise import Reader
from surprise import Dataset, evaluate

import time
start_time = time.time()

from pymongo import MongoClient
client = MongoClient("mongodb://notadmin:notpassword@ds117271.mlab.com:17271/soundvis")
db = client.soundvis
listening_sessions_repository = db["listening_sessions"]
stations_repository = db["stations"]
cursor = listening_sessions_repository.find({})

# aggregate all of the user's sessions into a double dict, where key = user_id,
# value = dictionary where key = station_id, value = duration listened to that session
aggregated_sessions = {}
for document in cursor:
    user_id = document['userId']
    station_id = document['stationId']
    duration = document['duration']
    if user_id not in aggregated_sessions:
        aggregated_sessions[user_id] = {}
    if station_id in aggregated_sessions[user_id]:
        aggregated_sessions[user_id][station_id] += duration
    else:
        aggregated_sessions[user_id][station_id] = duration

    if 'total' in aggregated_sessions[user_id]:
        aggregated_sessions[user_id]['total'] += duration
    else:
        aggregated_sessions[user_id]['total'] = duration

def create_rating(duration, total_duration):
    # if listen for over 5 minutes, rating is min 3
    if duration > 60 * 5:
        rating = 3 + (duration / total_duration) * 2
    else:
        rating = 1 + (duration / total_duration) * 4
    return rating

from bson import ObjectId
def print_station_genre_and_rating(station_id, rating):
    station = stations_repository.find_one({"_id": ObjectId(station_id)})
    if station != None:
        if 'genre' in station:
            print(station['genre'], rating)
        else:
            print("Genre not found for ", station_id)

# print out all real user's ratings
print("Users values:")
for user_id, station_durations in aggregated_sessions.items():
    print("User_id: ", user_id)
    total = station_durations['total']
    for station_id, duration in station_durations.items():
        if station_id != 'total':
            rating = create_rating(duration, total)
            print_station_genre_and_rating(station_id, rating)
print("Done printing users values")

from random import *
def create_user_with_ratings(user_id, genres):
    cursor = stations_repository.find({"genre": {"$in": genres}})
    total = 0
    for document in cursor:
        if str(document["genre"] == ""):
            continue
        if user_id not in aggregated_sessions:
            aggregated_sessions[user_id] = {}
        station_id = str(document["_id"])
        # random duration between 5 mins and 10 hours
        random_duration = randint(60 * 5, 60 * 10)
        aggregated_sessions[user_id][station_id] = random_duration
        total += random_duration
    if user_id in aggregated_sessions:
        for document in stations_repository.find({}):
            station_id = str(document["_id"])
            if station_id not in aggregated_sessions[user_id]:
                # random duration between 0 and 5 mins
                random_duration = randint(0, 60 * 5)
                aggregated_sessions[user_id][station_id] = random_duration
                total += random_duration
        aggregated_sessions[user_id]['total'] = total

# create fake users that only listen to specific genre
STRIDE = 100
for x in range(0, STRIDE):
    create_user_with_ratings(x, ["80s"])
for x in range(STRIDE, 2*STRIDE):
    create_user_with_ratings(x, ["Top 40", "Pop"])
for x in range(3*STRIDE, 4*STRIDE):
    create_user_with_ratings(x, ["Dance", "Drum and Bass", "Electronic", "House"])
# for x in range(4*STRIDE, 5*STRIDE):
#     create_user_with_ratings(x, ["Oldies", "Old Time Radio", "60s", "70s"])
# for x in range(5 * STRIDE, 6 * STRIDE):
#     create_user_with_ratings(x, ["News", "Talk", "Comedy"])
# for x in range(6 * STRIDE, 7 * STRIDE):
#     create_user_with_ratings(x, ["Rock", "Classic Rock", "Electric Blues", "Hard Rock", "Blues"])
for x in range(7 * STRIDE, 8 * STRIDE):
    create_user_with_ratings(x, ["Country", "Hot Country Hits"])
# for x in range(8 * STRIDE, 9 * STRIDE):
#     create_user_with_ratings(x, ["JPOP", "Japanese", "Asian"])
# for x in range(9 * STRIDE, 10 * STRIDE):
#     create_user_with_ratings(x, ["Classical", "Baroque"])
for x in range(10 * STRIDE, 11 * STRIDE):
    create_user_with_ratings(x, ["Jazz", "Smooth Jazz"])
# for x in range(11 * STRIDE, 12 * STRIDE):
#     create_user_with_ratings(x, ["Indian", "Bollywood"])

ratings_dict = {'userID': [], 'itemID': [], 'rating': []}

def add_entry(user_id, station_id, rating):
    ratings_dict['userID'].append(user_id)
    ratings_dict['itemID'].append(station_id)
    ratings_dict['rating'].append(rating)

# iterate through aggregated sessions and format into ratings_dict
for user_id, station_durations in aggregated_sessions.items():
    # print("User_id: ", user_id)
    total = station_durations['total']
    for station_id, duration in station_durations.items():
        if station_id != 'total':
            rating = create_rating(duration, total)
            add_entry(user_id, station_id, rating)

# format data
df = pd.DataFrame(ratings_dict)
reader = Reader(line_format='user item rating', rating_scale=(1, 5))
data = Dataset.load_from_df(df[['userID', 'itemID', 'rating']], reader)

# train model
from surprise import SVD
algo = SVD()

trainset = data.build_full_trainset()
algo.fit(trainset)

# save model to file
import pickle
file_name = 'recommendations/model'
dump_obj = {'predictions': None,
            'algo': algo
            }
pickle.dump(dump_obj, open(file_name, 'wb'), protocol=2)

end_time = time.time()

print("Elapsed time: ", end_time - start_time)