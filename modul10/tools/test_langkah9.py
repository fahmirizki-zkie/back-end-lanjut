import urllib.request
import json
import ssl

BASE = 'http://localhost:3000/api/v1'

def post(url, data, token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    req = urllib.request.Request(url, data=json.dumps(data).encode(), headers=headers, method='POST')
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.status, json.loads(resp.read().decode()), resp.headers
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        try:
            return e.code, json.loads(body), e.headers
        except:
            return e.code, body, e.headers

def get(url, token=None):
    headers = {}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    req = urllib.request.Request(url, headers=headers, method='GET')
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.status, json.loads(resp.read().decode()), resp.headers
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        try:
            return e.code, json.loads(body), e.headers
        except:
            return e.code, body, e.headers

def patch(url, data, token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    req = urllib.request.Request(url, data=json.dumps(data).encode(), headers=headers, method='PATCH')
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.status, json.loads(resp.read().decode()), resp.headers
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        try:
            return e.code, json.loads(body), e.headers
        except:
            return e.code, body, e.headers

def delete(url, token=None):
    headers = {}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    req = urllib.request.Request(url, headers=headers, method='DELETE')
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return resp.status, resp.read().decode(), resp.headers
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        try:
            return e.code, json.loads(body), e.headers
        except:
            return e.code, body, e.headers

# 1. Login users (cache tokens to avoid rate limit)
import os
import time

TOKEN_CACHE = 'tools/.tokens.json'
tokens = {}
if os.path.exists(TOKEN_CACHE):
    try:
        with open(TOKEN_CACHE, 'r') as f:
            tokens = json.load(f)
    except:
        tokens = {}

needed = [u for u in ['admin', 'mhs001', 'mhs002', 'mhs003'] if u not in tokens]
if needed:
    print(f"Logging in users: {needed}")
    for u in needed:
        st, res, _ = post(f'{BASE}/auth/login', {'username': u, 'password': 'password'})
        if st == 200:
            tokens[u] = res['data']['access_token']
            print(f'Login {u}: OK')
        elif st == 429:
            print(f'Login {u} hit rate limit. Waiting 60s...')
            time.sleep(60)
            st, res, _ = post(f'{BASE}/auth/login', {'username': u, 'password': 'password'})
            if st == 200:
                tokens[u] = res['data']['access_token']
                print(f'Login {u}: OK (after wait)')
            else:
                print(f'Login {u} failed: {st} {res}')
        else:
            print(f'Login {u} failed: {st} {res}')
    with open(TOKEN_CACHE, 'w') as f:
        json.dump(tokens, f)
else:
    print("Using cached tokens for all users.")

# Clean up any existing reviews on course 1 to make test idempotent
st, res, _ = get(f'{BASE}/courses/1/reviews', tokens['admin'])
if st == 200 and 'data' in res:
    for r in res['data']:
        delete(f'{BASE}/reviews/{r["id"]}', tokens['admin'])
    print(f"Cleaned up {len(res['data'])} existing reviews on course 1.")

# Test 1: mhs003 writes review IF101 -> 403 NOT_ENROLLED
st, res, _ = post(f'{BASE}/courses/1/reviews', {'rating': 5, 'comment': 'Materi Go mantap sekali', 'tags': ['seru']}, tokens['mhs003'])
print('Test 1 (mhs003 NOT_ENROLLED):', st, res.get('code') if isinstance(res, dict) else res)

# Test 2: mhs001 writes review with dirty tags -> 201 Created
st, res, hdr = post(f'{BASE}/courses/1/reviews', {
    'rating': 4,
    'comment': 'Materi Go-nya jelas, tugasnya cukup banyak',
    'tags': ['Tugas-Banyak', ' dosen-jelas ', 'tugas-banyak']
}, tokens['mhs001'])
print('Test 2 (mhs001 201 OK):', st, 'tags:', res['data']['tags'], 'Location:', hdr.get('Location'))
review_id = res['data']['id']

# Test 3: mhs001 writes second review -> 409 ALREADY_REVIEWED
st, res, _ = post(f'{BASE}/courses/1/reviews', {'rating': 5, 'comment': 'Coba tulis ulasan kedua deh', 'tags': ['seru']}, tokens['mhs001'])
print('Test 3 (mhs001 duplicate):', st, res.get('code') if isinstance(res, dict) else res)

# Test 4: Validation rating 9, pendek, 6 tags -> 422
st, res, _ = post(f'{BASE}/courses/1/reviews', {'rating': 9, 'comment': 'pendek', 'tags': ['1','2','3','4','5','6']}, tokens['mhs001'])
print('Test 4 (Validation 422):', st, res.get('errors') if isinstance(res, dict) else res)

# Test 5: Course 99999 -> 404
st, res, _ = post(f'{BASE}/courses/99999/reviews', {'rating': 5, 'comment': 'Kelas tidak ada di sistem', 'tags': ['oke']}, tokens['mhs001'])
print('Test 5 (Course 99999):', st, res.get('code') if isinstance(res, dict) else res)

# Test 6: GET /reviews/bukan-objectid -> 400
st, res, _ = get(f'{BASE}/reviews/bukan-objectid', tokens['mhs001'])
print('Test 6 (Invalid ObjectId):', st, res.get('message') if isinstance(res, dict) else res)

# Test 7: mhs002 edits mhs001 review -> 403
st, res, _ = patch(f'{BASE}/reviews/{review_id}', {'rating': 3}, tokens['mhs002'])
print('Test 7 (Edit other review):', st, res.get('code') if isinstance(res, dict) else res)

# Test 8: mhs001 edits own review -> 200 OK
st, res, _ = patch(f'{BASE}/reviews/{review_id}', {'comment': 'Materi Go-nya sangat jelas sekali dan menantang'}, tokens['mhs001'])
print('Test 8 (Edit own review):', st, 'updated_at:', res['data']['updated_at'])

# Test 9: mhs001 marks own review helpful -> 409 OWN_REVIEW
st, res, _ = post(f'{BASE}/reviews/{review_id}/helpful', {}, tokens['mhs001'])
print('Test 9 (Vote own review):', st, res.get('code') if isinstance(res, dict) else res)

# Test 10: mhs002 marks helpful -> 200, then repeats -> 409 ALREADY_VOTED
st1, res1, _ = post(f'{BASE}/reviews/{review_id}/helpful', {}, tokens['mhs002'])
st2, res2, _ = post(f'{BASE}/reviews/{review_id}/helpful', {}, tokens['mhs002'])
print('Test 10 (Helpful):', st1, 'count:', res1['data']['helpful_count'], 'Repeat:', st2, res2.get('code') if isinstance(res2, dict) else res2)

# Test 11: mhs003 deletes mhs001 review -> 403
st, res, _ = delete(f'{BASE}/reviews/{review_id}', tokens['mhs003'])
print('Test 11 (mhs003 delete):', st, res.get('code') if isinstance(res, dict) else res)

# Test 12: admin deletes review -> 204
st, res, _ = delete(f'{BASE}/reviews/{review_id}', tokens['admin'])
print('Test 12 (Admin delete):', st)

# Test 13: Delete already deleted -> 404
st, res, _ = delete(f'{BASE}/reviews/{review_id}', tokens['admin'])
print('Test 13 (Delete again):', st, res.get('code') if isinstance(res, dict) else res)

# Test 14: NoSQL injection in body -> 400
raw_body = b'{"comment": {"$ne": ""}}'
req = urllib.request.Request(
    f'{BASE}/courses/1/reviews',
    data=raw_body,
    headers={'Content-Type': 'application/json', 'Authorization': f'Bearer {tokens["mhs001"]}'},
    method='POST'
)
try:
    with urllib.request.urlopen(req) as resp:
        print('Test 14 (NoSQL body):', resp.status)
except urllib.error.HTTPError as e:
    body = e.read().decode()
    msg = json.loads(body).get('message') if e.headers.get_content_type() == 'application/json' else body
    print('Test 14 (NoSQL body):', e.code, msg)

# Test 15: NoSQL injection in query param -> 200 OK with 0 items
st, res, _ = get(f'{BASE}/courses/1/reviews?tag=%7B%22$ne%22:%22%22%7D', tokens['mhs001'])
print('Test 15 (NoSQL query):', st, 'count data:', len(res['data']))
