import type { Device } from "~/lib/model";
import Instructions from "./Instructions";

interface Props {
	device: Device;
}

export default function LinuxInstructions(props: Props) {
	return (
		<>
			<Instructions.Title>
				Configure DNS over TLS using{" "}
				<span class="font-medium tracking-wide">systemd-resolved</span> to
				enable blkhole.
			</Instructions.Title>
			<Instructions>
				<li>
					Open the configuration file:
					<div class="mx-4 mt-2 mb-6 font-medium text-xs tracking-wide">
						/etc/systemd/resolved.conf
					</div>
				</li>
				<li>
					Add the following:
					<div class="mx-4 mt-2 mb-6 font-medium text-xs tracking-wide">
						[Resolve]
						<br />
						{`DNS=${props.device.hash.toLowerCase()}.blkhole.sh`}
						<br />
						DNSOverTLS=yes
					</div>
				</li>
				<li>
					Save the file and restart the service:
					<div class="mx-4 mt-2 mb-6 font-medium text-xs tracking-wide">
						sudo systemctl restart systemd-resolved
					</div>
				</li>
			</Instructions>
		</>
	);
}
