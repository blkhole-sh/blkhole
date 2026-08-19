export interface PairResponse {
	accessToken: string;
	client: {
		id: number;
		name: string;
		browser?: string;
		createdAt: string;
		lastActiveAt: string;
	};
	device: {
		id: number;
		name: string;
	};
}

export interface ExtensionState {
	apiOrigin: string;
	accessToken?: string;
	etag?: string;
	client?: PairResponse["client"];
	device?: PairResponse["device"];
	domainCount?: number;
	lastSyncedAt?: string;
}

export interface DynamicRule {
	id: number;
	priority: number;
	action: {
		type: "redirect";
		redirect: { extensionPath: string };
	};
	condition: {
		requestDomains: string[];
		resourceTypes: ["main_frame"];
	};
}

export type BridgeRequest =
	| {
			source: "blkhole-webapp";
			type: "BLKHOLE_EXTENSION_PING";
			nonce: string;
	  }
	| {
			source: "blkhole-webapp";
			type: "BLKHOLE_EXTENSION_PAIR";
			nonce: string;
			pairingToken: string;
			apiOrigin: string;
	  };
