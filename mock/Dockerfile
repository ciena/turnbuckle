# Copyright 2021 Ciena Corporation.
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
FROM python:3.7

# Create a user, so we don't run as root
RUN adduser -q --shell /bin/bash --disabled-login --gecos "" mock
USER mock
WORKDIR /home/mock
ENV PATH="/home/mock/.local/bin:${PATH}"

# Add the source and install the requirements
ADD . /src
RUN pip install -r /src/requirements.txt

# And off we go ...
CMD kopf run --verbose --debug --all-namespaces /src/controller.py
