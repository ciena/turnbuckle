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

import kopf
import os
import sys
import grpc
import re
from concurrent import futures
import threading
from apis import ruleprovider_pb2_grpc
from apis import ruleprovider_pb2
import logging
import random
import time
from grpc_reflection.v1alpha import reflection
from kubernetes import client as k8s_client, config as k8s_config


class RuleProviderServicer(ruleprovider_pb2_grpc.RuleProviderServicer):
    def Evaluate(self, request, context):
        logging.getLogger()
        try:
            api = k8s_client.CustomObjectsApi()
            ret = api.list_cluster_custom_object('constraint.ciena.com', 'v1',
                                                 'ruleproviders')

            # sort the list of returned items based on priority, high to low
            ret['items'].sort(key=lambda x: x['spec']['priority'], reverse=True)

            # Walk the list of returned items and see if we have a match for
            # the request
            for item in ret['items']:

                # if the rule name doesn't match the rule provider or if there
                # is a match in the number of targets, then no match, so skip
                if (
                        item['spec']['rule'] != request.rule.name or
                        len(item['spec']['targets']) != len(request.targets)
                     ):
                    continue

                # convert item targets to dict
                targets = {}
                for target in item['spec']['targets']:
                    ref = target['reference']
                    parts = ref.split(":")
                    count = len(parts)
                    if count == 1:  # apiVersion, kind, name
                        ref = '.*:.*:.*:.*:{}'.format(parts[0])
                    elif count == 2:  # apiVersion, kind, name
                        ref = '.*:.*:.*:{}:{}'.format(parts[0],
                                                      parts[1])
                    elif count == 3:  # apiVersion, kind, name
                        ref = '.*:.*:{}:{}:{}'.format(parts[0],
                                                      parts[1],
                                                      parts[2])
                    elif count == 4:  # ns, apiVersion, kind, name
                        ref = '.*:{}:{}:{}:{}'.format(parts[0],
                                                      parts[1],
                                                      parts[2],
                                                      parts[3])

                    targets[target['name']] = ref
                    logging.debug('{} = {}'.format(target['name'], ref))

                # Compare the target lists and if they are the same then this
                # entry can be used
                match = True
                for key, val in request.targets.items():

                    # convert value to single string
                    ref = '{}:{}:{}:{}:{}'.format(val.cluster,
                                                  val.namespace,
                                                  val.apiVersion,
                                                  val.kind,
                                                  val.name)
                    if key not in targets or re.match(targets[key], ref) is None:
                        match = False
                        logging.debug('no match found for {} = {}'
                                      .format(key, ref))
                        break

                if match:
                    logging.debug('provider match found, returning: {} => {}'
                                  .format(item['spec']['value'],
                                          item['spec']['reason']))
                    return ruleprovider_pb2.EvaluateResponse(reason=item['spec']['reason'],
                                                             compliance=item['spec']['value'])

            compliance = os.getenv('DEFAULT_COMPLIANCE', 'Compliant')
            reason = os.getenv('DEFAULT_REASON', 'no value configured')
            logging.debug('provider not found, using default')
            return ruleprovider_pb2.EvaluateResponse(reason=reason,
                                                     compliance=compliance)
        except Exception as ex:
            logging.error("unable to query rule providers: {}".format(ex))
            return ruleprovider_pb2.EvaluationResponse(compliance='Compliant',
                                                       reason='unable to access k8s API')
    def EndpointCost(self, request, context):
        logging.getLogger()
        nc = []
        for node in request.eligibleNodes:
            cost = random.randint(1, 10000)
            nc.append(ruleprovider_pb2.NodeCost(node=node, cost=cost))
            logger.debug('node {}, assigned cost {}'.format(node, cost))

        return ruleprovider_pb2.EndpointCostResponse(nodeAndCost = nc)

random.seed(time.time())
logging.getLogger()
server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
ruleprovider_pb2_grpc.add_RuleProviderServicer_to_server(
    RuleProviderServicer(), server)
SERVICE_NAMES = (
    ruleprovider_pb2.DESCRIPTOR.services_by_name['RuleProvider'].full_name,
    reflection.SERVICE_NAME,
)
logging.debug(SERVICE_NAMES)
reflection.enable_server_reflection(SERVICE_NAMES, server)

logging.debug("listening on 0.0.0.0:5309 ...")
server.add_insecure_port('0.0.0.0:5309')
server.start()

# First try in cluster config
try:
    k8s_config.load_incluster_config()
except Exception:
    try:
        k8s_config.load_kube_config()
    except Exception:
        logging.error("unable to find k8s configuration")
        sys.exit(1)
