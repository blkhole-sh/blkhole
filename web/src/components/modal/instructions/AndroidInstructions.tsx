import type { Device } from "~/lib/model";
import Instructions from "./Instructions";

interface Props {
	device: Device;
}

export default function AndroidInstructions(props: Props) {
	return (
		<>
			<Instructions.Title>
				Configure Private DNS to enable blkhole.
			</Instructions.Title>
			<Instructions>
				<li>
					Go to{" "}
					<span class="font-medium tracking-wide">
						Settings → Network & Internet → Private DNS
					</span>
				</li>
				<li>
					Select{" "}
					<span class="font-medium tracking-wide">
						Private DNS provider hostname
					</span>
				</li>
				<li>
					Enter:{" "}
					<span class="font-medium tracking-wide">
						{props.device.hash.toLowerCase()}.blkhole.sh
					</span>
				</li>
				<li>
					Tap <span class="font-medium tracking-wide">Save</span>
				</li>
			</Instructions>
		</>
	);
}
