/**
 * Autogenerated by Frugal Compiler (2.0.0-RC8)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package variety.java;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.FrugalTApplicationExceptionType;
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.*;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.transport.FPublisherTransport;
import com.workiva.frugal.transport.FSubscriberTransport;
import com.workiva.frugal.transport.FSubscription;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.thrift.TException;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.protocol.*;

import java.util.Arrays;
import java.util.List;
import java.util.logging.Logger;
import javax.annotation.Generated;




@Generated(value = "Autogenerated by Frugal Compiler (2.0.0-RC8)", date = "2015-11-24")
public class EventsSubscriber {

	/**
	 * This docstring gets added to the generated code because it has
	 * the @ sign. Prefix specifies topic prefix tokens, which can be static or
	 * variable.
	 */
	public interface Iface {
		/**
		 * This is a docstring.
		 */
		public FSubscription subscribeEventCreated(String user, final EventCreatedHandler handler) throws TException;

	}

	public interface EventCreatedHandler {
		void onEventCreated(FContext ctx, Event req);
	}

	/**
	 * This docstring gets added to the generated code because it has
	 * the @ sign. Prefix specifies topic prefix tokens, which can be static or
	 * variable.
	 */
	public static class Client implements Iface {
		private static final String DELIMITER = ".";
		private static final Logger LOGGER = Logger.getLogger(Client.class.getName());

		private final FScopeProvider provider;
		private final ServiceMiddleware[] middleware;

		public Client(FScopeProvider provider, ServiceMiddleware... middleware) {
			this.provider = provider;
			List<ServiceMiddleware> combined = Arrays.asList(middleware);
			combined.addAll(provider.getMiddleware());
			this.middleware = combined.toArray(new ServiceMiddleware[0]);
		}

		/**
		 * This is a docstring.
		 */
		public FSubscription subscribeEventCreated(String user, final EventCreatedHandler handler) throws TException {
			final String op = "EventCreated";
			String prefix = String.format("foo.%s.", user);
			final String topic = String.format("%sEvents%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final EventCreatedHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, EventCreatedHandler.class, middleware);
			transport.subscribe(topic, recvEventCreated(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvEventCreated(String op, FProtocolFactory pf, EventCreatedHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(FrugalTApplicationExceptionType.UNKNOWN_METHOD);
					}
					Event received = new Event();
					received.read(iprot);
					iprot.readMessageEnd();
					handler.onEventCreated(ctx, received);
				}
			};
		}

	}

}
