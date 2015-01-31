#!/bin/bash
curl -i -H "Content-Type: application/json" -d '{"name":"New Account", "description":"Description of the account"}' localhost:8082/accounts
curl -i -H "Content-Type: application/json" -d '{"user_name":"User name", "password":"hmac(256)", "email":"de@m.o"}' localhost:8082/account/1/users
curl -i -H "Content-Type: application/json" -d '{"key": "hmac(256)", "name":"New App"}' localhost:8082/account/1/applications
curl -i -H "Content-Type: application/json" -d '{"auth_token": "yZg6ZCJjHGy5caTcVnD25pVMEswUEQWTSA64tkBU", "username": "dlsniper", "name": "Florin Patan", "password": "JjQxWYYCWfX634q6KQeDSusywVH3T5Dw9hMqBcUd", "email": "florinpatan@gmail.com", "url": "http://florinpatan.ro", "metadata": "{}"}' localhost:8082/application/1/users
curl -i -H "Content-Type: application/json" -d '{"auth_token": "yZg6ZCJjHGy5caTcVnD25pVMEswUEQWTSA64tkBU", "username": "dlsniper", "name": "Florin Patan", "password": "JjQxWYYCWfX634q6KQeDSusywVH3T5Dw9hMqBcUd", "email": "florinpatan@gmail.com", "url": "http://florinpatan.ro", "metadata": "{}"}' localhost:8082/application/1/users
curl -i -H "Content-Type: application/json" -d '{"auth_token": "yZg6ZCJjHGy5caTcVnD25pVMEswUEQWTSA64tkBU", "username": "dlsniper", "name": "Florin Patan", "password": "JjQxWYYCWfX634q6KQeDSusywVH3T5Dw9hMqBcUd", "email": "florinpatan@gmail.com", "url": "http://florinpatan.ro", "metadata": "{}"}' localhost:8082/application/1/users
curl -i -H "Content-Type: application/json" -d '{"auth_token": "yZg6ZCJjHGy5caTcVnD25pVMEswUEQWTSA64tkBU", "username": "dlsniper", "name": "Florin Patan", "password": "JjQxWYYCWfX634q6KQeDSusywVH3T5Dw9hMqBcUd", "email": "florinpatan@gmail.com", "url": "http://florinpatan.ro", "metadata": "{}"}' localhost:8082/application/1/users
curl -i -H "Content-Type: application/json" -d '{"auth_token": "yZg6ZCJjHGy5caTcVnD25pVMEswUEQWTSA64tkBU", "username": "dlsniper", "name": "Florin Patan", "password": "JjQxWYYCWfX634q6KQeDSusywVH3T5Dw9hMqBcUd", "email": "florinpatan@gmail.com", "url": "http://florinpatan.ro", "metadata": "{}"}' localhost:8082/application/1/users
curl -i -H "Content-Type: application/json" -d '{"user_from_id":1,"user_to_id":2}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":1,"user_to_id":3}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":1,"user_to_id":5}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":2,"user_to_id":5}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":2,"user_to_id":3}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":3,"user_to_id":5}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"user_from_id":4,"user_to_id":5}' localhost:8082/application/1/connections
curl -i -H "Content-Type: application/json" -d '{"verb": "likese;uoiyusyhuot8woityzitup8w4t4wiutziu[aw84yt4ieutzw4yt[5yh4twy4rt[8wu56iu[8 s375[3s6sou6;s876seuy/xe49uys03755rut.ty", "metadata": "{}"}' localhost:8082/application/1/user/1/events
curl -i -H "Content-Type: application/json" -d '{"verb": "likese;uoiyusyhuot8woityzitup8w4t4wiutziu[aw84yt4ieutzw4yt[5yh4twy4rt[8wu56iu[8 s375[3s6sou6;s876seuy/xe49uys03755rut.ty", "metadata": "{}"}' localhost:8082/application/1/user/2/events
curl -i -H "Content-Type: application/json" -d '{"verb": "likese;uoiyusyhuot8woityzitup8w4t4wiutziu[aw84yt4ieutzw4yt[5yh4twy4rt[8wu56iu[8 s375[3s6sou6;s876seuy/xe49uys03755rut.ty", "metadata": "{}"}' localhost:8082/application/1/user/3/events
curl -i -H "Content-Type: application/json" -d '{"verb": "likese;uoiyusyhuot8woityzitup8w4t4wiutziu[aw84yt4ieutzw4yt[5yh4twy4rt[8wu56iu[8 s375[3s6sou6;s876seuy/xe49uys03755rut.ty", "metadata": "{}"}' localhost:8082/application/1/user/4/events
curl -i -H "Content-Type: application/json" -d '{"verb": "likese;uoiyusyhuot8woityzitup8w4t4wiutziu[aw84yt4ieutzw4yt[5yh4twy4rt[8wu56iu[8 s375[3s6sou6;s876seuy/xe49uys03755rut.ty", "metadata": "{}"}' localhost:8082/application/1/user/5/events
