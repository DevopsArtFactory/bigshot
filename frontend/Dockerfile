# Copyright 2020 The bigshot Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM node:12.19.0-stretch as node
WORKDIR /app/modules
COPY package.json /app/modules
RUN npm install
RUN npm install node-sass

WORKDIR /app
NV NODE_PATH /app/modules/node_modules
