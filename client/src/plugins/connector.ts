import Vue from 'vue';
import { grpc } from '@improbable-eng/grpc-web';
import * as jspb from 'google-protobuf';
import * as google_protobuf_empty_pb from 'google-protobuf/google/protobuf/empty_pb';
import { Empty } from 'google-protobuf/google/protobuf/empty_pb';

import { Features } from '../rpc/protocol/features_pb_service';

import {
	Profile,
	UpdateProfileRequest,
	ThermalResponse
} from '../rpc/protocol/thermal_pb'

import {
	Thermal,
	ThermalGetCurrentProfile
} from '../rpc/protocol/thermal_pb_service'


class Connector {
	constructor() {
		const profile = new Profile();
		// const req = new UpdateProfileRequest();
		const req = new Empty;

		grpc.unary(Thermal.GetCurrentProfile, {
			request: req,
			host: 'http://127.0.0.1:41959',
			onEnd: (res) => {
				const { status, statusMessage, headers, message, trailers } = res;
				console.log('onEnd.status', status, statusMessage);
				console.log('onEnd.headers', headers);

				if (status === grpc.Code.OK && message) {
					console.log('YAS!');
					const resp = ThermalResponse.deserializeBinary(message.serializeBinary());
					console.log(resp);
					// console.log(resp.);
					//   console.log('onEnd.message', message.toObject());
					// const resp = FeaturesResponse.deserializeBinary(
					// 	message.serializeBinary()
					// );
					// console.log('success', resp.getSuccess());
					// console.log('rog', resp.getFeature().getRogremapList());
				}

				console.log('onEnd.trailers', trailers);
			},
		})
		// const fanCurve = profile.getCpufancurve();

		// console.log('fanCurve', fanCurve);
	}

	// async grpcRequest() {
	// }
}


// Vue.use({
// 	install: (Vue, options) => {
// 		const connctorInstance = new Connector();
// 		Vue.prototype.$connector = connctorInstance;
// 	}
// });

export default {
	install: () => {
		const connctorInstance = new Connector();
		Vue.prototype.$connector = connctorInstance;
	}
}
