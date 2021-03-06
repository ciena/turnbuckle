# Copyright 2022 Ciena Corporation.
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

# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: apis/ruleprovider.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='apis/ruleprovider.proto',
  package='ruleprovider',
  syntax='proto3',
  serialized_options=b'Z\021apis/ruleprovider',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x17\x61pis/ruleprovider.proto\x12\x0cruleprovider\":\n\nPolicyRule\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\x0f\n\x07request\x18\x02 \x01(\t\x12\r\n\x05limit\x18\x03 \x01(\t\"\\\n\x06Target\x12\x0f\n\x07\x63luster\x18\x01 \x01(\t\x12\x11\n\tnamespace\x18\x02 \x01(\t\x12\x12\n\napiVersion\x18\x03 \x01(\t\x12\x0c\n\x04kind\x18\x04 \x01(\t\x12\x0c\n\x04name\x18\x05 \x01(\t\"\xbc\x01\n\x0f\x45valuateRequest\x12;\n\x07targets\x18\x01 \x03(\x0b\x32*.ruleprovider.EvaluateRequest.TargetsEntry\x12&\n\x04rule\x18\x02 \x01(\x0b\x32\x18.ruleprovider.PolicyRule\x1a\x44\n\x0cTargetsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12#\n\x05value\x18\x02 \x01(\x0b\x32\x14.ruleprovider.Target:\x02\x38\x01\"\xba\x01\n\x10\x45valuateResponse\x12\x42\n\ncompliance\x18\x01 \x01(\x0e\x32..ruleprovider.EvaluateResponse.ComplianceState\x12\x0e\n\x06reason\x18\x02 \x01(\t\"R\n\x0f\x43omplianceState\x12\x0b\n\x07Pending\x10\x00\x12\r\n\tCompliant\x10\x01\x12\t\n\x05Limit\x10\x02\x12\r\n\tViolation\x10\x03\x12\t\n\x05\x45rror\x10\x04\"&\n\x08NodeCost\x12\x0c\n\x04node\x18\x01 \x01(\t\x12\x0c\n\x04\x63ost\x18\x02 \x01(\x03\"C\n\x14\x45ndpointCostResponse\x12+\n\x0bnodeAndCost\x18\x01 \x03(\x0b\x32\x16.ruleprovider.NodeCost\"\x8d\x01\n\x13\x45ndpointCostRequest\x12$\n\x06source\x18\x01 \x01(\x0b\x32\x14.ruleprovider.Target\x12&\n\x04rule\x18\x02 \x01(\x0b\x32\x18.ruleprovider.PolicyRule\x12\x11\n\tpeerNodes\x18\x03 \x03(\t\x12\x15\n\religibleNodes\x18\x04 \x03(\t2\xb0\x01\n\x0cRuleProvider\x12I\n\x08\x45valuate\x12\x1d.ruleprovider.EvaluateRequest\x1a\x1e.ruleprovider.EvaluateResponse\x12U\n\x0c\x45ndpointCost\x12!.ruleprovider.EndpointCostRequest\x1a\".ruleprovider.EndpointCostResponseB\x13Z\x11\x61pis/ruleproviderb\x06proto3'
)



_EVALUATERESPONSE_COMPLIANCESTATE = _descriptor.EnumDescriptor(
  name='ComplianceState',
  full_name='ruleprovider.EvaluateResponse.ComplianceState',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='Pending', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='Compliant', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='Limit', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='Violation', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='Error', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=491,
  serialized_end=573,
)
_sym_db.RegisterEnumDescriptor(_EVALUATERESPONSE_COMPLIANCESTATE)


_POLICYRULE = _descriptor.Descriptor(
  name='PolicyRule',
  full_name='ruleprovider.PolicyRule',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='ruleprovider.PolicyRule.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='request', full_name='ruleprovider.PolicyRule.request', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='limit', full_name='ruleprovider.PolicyRule.limit', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=41,
  serialized_end=99,
)


_TARGET = _descriptor.Descriptor(
  name='Target',
  full_name='ruleprovider.Target',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='cluster', full_name='ruleprovider.Target.cluster', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='namespace', full_name='ruleprovider.Target.namespace', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='apiVersion', full_name='ruleprovider.Target.apiVersion', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='kind', full_name='ruleprovider.Target.kind', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='name', full_name='ruleprovider.Target.name', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=101,
  serialized_end=193,
)


_EVALUATEREQUEST_TARGETSENTRY = _descriptor.Descriptor(
  name='TargetsEntry',
  full_name='ruleprovider.EvaluateRequest.TargetsEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='ruleprovider.EvaluateRequest.TargetsEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='ruleprovider.EvaluateRequest.TargetsEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=b'8\001',
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=316,
  serialized_end=384,
)

_EVALUATEREQUEST = _descriptor.Descriptor(
  name='EvaluateRequest',
  full_name='ruleprovider.EvaluateRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='targets', full_name='ruleprovider.EvaluateRequest.targets', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='rule', full_name='ruleprovider.EvaluateRequest.rule', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_EVALUATEREQUEST_TARGETSENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=196,
  serialized_end=384,
)


_EVALUATERESPONSE = _descriptor.Descriptor(
  name='EvaluateResponse',
  full_name='ruleprovider.EvaluateResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='compliance', full_name='ruleprovider.EvaluateResponse.compliance', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='reason', full_name='ruleprovider.EvaluateResponse.reason', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
    _EVALUATERESPONSE_COMPLIANCESTATE,
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=387,
  serialized_end=573,
)


_NODECOST = _descriptor.Descriptor(
  name='NodeCost',
  full_name='ruleprovider.NodeCost',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='node', full_name='ruleprovider.NodeCost.node', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='cost', full_name='ruleprovider.NodeCost.cost', index=1,
      number=2, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=575,
  serialized_end=613,
)


_ENDPOINTCOSTRESPONSE = _descriptor.Descriptor(
  name='EndpointCostResponse',
  full_name='ruleprovider.EndpointCostResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='nodeAndCost', full_name='ruleprovider.EndpointCostResponse.nodeAndCost', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=615,
  serialized_end=682,
)


_ENDPOINTCOSTREQUEST = _descriptor.Descriptor(
  name='EndpointCostRequest',
  full_name='ruleprovider.EndpointCostRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='source', full_name='ruleprovider.EndpointCostRequest.source', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='rule', full_name='ruleprovider.EndpointCostRequest.rule', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='peerNodes', full_name='ruleprovider.EndpointCostRequest.peerNodes', index=2,
      number=3, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='eligibleNodes', full_name='ruleprovider.EndpointCostRequest.eligibleNodes', index=3,
      number=4, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=685,
  serialized_end=826,
)

_EVALUATEREQUEST_TARGETSENTRY.fields_by_name['value'].message_type = _TARGET
_EVALUATEREQUEST_TARGETSENTRY.containing_type = _EVALUATEREQUEST
_EVALUATEREQUEST.fields_by_name['targets'].message_type = _EVALUATEREQUEST_TARGETSENTRY
_EVALUATEREQUEST.fields_by_name['rule'].message_type = _POLICYRULE
_EVALUATERESPONSE.fields_by_name['compliance'].enum_type = _EVALUATERESPONSE_COMPLIANCESTATE
_EVALUATERESPONSE_COMPLIANCESTATE.containing_type = _EVALUATERESPONSE
_ENDPOINTCOSTRESPONSE.fields_by_name['nodeAndCost'].message_type = _NODECOST
_ENDPOINTCOSTREQUEST.fields_by_name['source'].message_type = _TARGET
_ENDPOINTCOSTREQUEST.fields_by_name['rule'].message_type = _POLICYRULE
DESCRIPTOR.message_types_by_name['PolicyRule'] = _POLICYRULE
DESCRIPTOR.message_types_by_name['Target'] = _TARGET
DESCRIPTOR.message_types_by_name['EvaluateRequest'] = _EVALUATEREQUEST
DESCRIPTOR.message_types_by_name['EvaluateResponse'] = _EVALUATERESPONSE
DESCRIPTOR.message_types_by_name['NodeCost'] = _NODECOST
DESCRIPTOR.message_types_by_name['EndpointCostResponse'] = _ENDPOINTCOSTRESPONSE
DESCRIPTOR.message_types_by_name['EndpointCostRequest'] = _ENDPOINTCOSTREQUEST
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

PolicyRule = _reflection.GeneratedProtocolMessageType('PolicyRule', (_message.Message,), {
  'DESCRIPTOR' : _POLICYRULE,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.PolicyRule)
  })
_sym_db.RegisterMessage(PolicyRule)

Target = _reflection.GeneratedProtocolMessageType('Target', (_message.Message,), {
  'DESCRIPTOR' : _TARGET,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.Target)
  })
_sym_db.RegisterMessage(Target)

EvaluateRequest = _reflection.GeneratedProtocolMessageType('EvaluateRequest', (_message.Message,), {

  'TargetsEntry' : _reflection.GeneratedProtocolMessageType('TargetsEntry', (_message.Message,), {
    'DESCRIPTOR' : _EVALUATEREQUEST_TARGETSENTRY,
    '__module__' : 'apis.ruleprovider_pb2'
    # @@protoc_insertion_point(class_scope:ruleprovider.EvaluateRequest.TargetsEntry)
    })
  ,
  'DESCRIPTOR' : _EVALUATEREQUEST,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.EvaluateRequest)
  })
_sym_db.RegisterMessage(EvaluateRequest)
_sym_db.RegisterMessage(EvaluateRequest.TargetsEntry)

EvaluateResponse = _reflection.GeneratedProtocolMessageType('EvaluateResponse', (_message.Message,), {
  'DESCRIPTOR' : _EVALUATERESPONSE,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.EvaluateResponse)
  })
_sym_db.RegisterMessage(EvaluateResponse)

NodeCost = _reflection.GeneratedProtocolMessageType('NodeCost', (_message.Message,), {
  'DESCRIPTOR' : _NODECOST,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.NodeCost)
  })
_sym_db.RegisterMessage(NodeCost)

EndpointCostResponse = _reflection.GeneratedProtocolMessageType('EndpointCostResponse', (_message.Message,), {
  'DESCRIPTOR' : _ENDPOINTCOSTRESPONSE,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.EndpointCostResponse)
  })
_sym_db.RegisterMessage(EndpointCostResponse)

EndpointCostRequest = _reflection.GeneratedProtocolMessageType('EndpointCostRequest', (_message.Message,), {
  'DESCRIPTOR' : _ENDPOINTCOSTREQUEST,
  '__module__' : 'apis.ruleprovider_pb2'
  # @@protoc_insertion_point(class_scope:ruleprovider.EndpointCostRequest)
  })
_sym_db.RegisterMessage(EndpointCostRequest)


DESCRIPTOR._options = None
_EVALUATEREQUEST_TARGETSENTRY._options = None

_RULEPROVIDER = _descriptor.ServiceDescriptor(
  name='RuleProvider',
  full_name='ruleprovider.RuleProvider',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=829,
  serialized_end=1005,
  methods=[
  _descriptor.MethodDescriptor(
    name='Evaluate',
    full_name='ruleprovider.RuleProvider.Evaluate',
    index=0,
    containing_service=None,
    input_type=_EVALUATEREQUEST,
    output_type=_EVALUATERESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='EndpointCost',
    full_name='ruleprovider.RuleProvider.EndpointCost',
    index=1,
    containing_service=None,
    input_type=_ENDPOINTCOSTREQUEST,
    output_type=_ENDPOINTCOSTRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_RULEPROVIDER)

DESCRIPTOR.services_by_name['RuleProvider'] = _RULEPROVIDER

# @@protoc_insertion_point(module_scope)
