#
# Autogenerated by Frugal Compiler (2.0.0-RC6)
#
# DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
#



import sys
import traceback

from thrift.Thrift import TApplicationException
from thrift.Thrift import TMessageType
from thrift.Thrift import TType
from tornado import gen
from frugal.middleware import Method
from frugal.subscription import FSubscription
from frugal.transport import TMemoryOutputBuffer

from .ttypes import *




class EventsPublisher(object):
    """
    This docstring gets added to the generated code because it has
    the @ sign. Prefix specifies topic prefix tokens, which can be static or
    variable.
    """

    _DELIMITER = '.'

    def __init__(self, provider, middleware=None):
        """
        Create a new EventsPublisher.

        Args:
            provider: FScopeProvider
            middleware: ServiceMiddleware or list of ServiceMiddleware
        """

        middleware = middleware or []
        if middleware and not isinstance(middleware, list):
            middleware = [middleware]
        middleware += provider.get_middleware()
        self._transport, self._protocol_factory = provider.new_publisher()
        self._methods = {
            'publish_EventCreated': Method(self._publish_EventCreated, middleware),
        }

    @gen.coroutine
    def open(self):
        yield self._transport.open()

    @gen.coroutine
    def close(self):
        yield self._transport.close()

    @gen.coroutine
    def publish_EventCreated(self, ctx, user, req):
        """
        This is a docstring.
        
        Args:
            ctx: FContext
            user: string
            req: Event
        """
        yield self._methods['publish_EventCreated']([ctx, user, req])

    @gen.coroutine
    def _publish_EventCreated(self, ctx, user, req):
        op = 'EventCreated'
        prefix = 'foo.{}.'.format(user)
        topic = '{}Events{}{}'.format(prefix, self._DELIMITER, op)
        buffer = TMemoryOutputBuffer(self._transport.get_publish_size_limit())
        oprot = self._protocol_factory.get_protocol(buffer)
        oprot.write_request_headers(ctx)
        oprot.writeMessageBegin(op, TMessageType.CALL, 0)
        req.write(oprot)
        oprot.writeMessageEnd()
        yield self._transport.publish(topic, buffer.getvalue())
