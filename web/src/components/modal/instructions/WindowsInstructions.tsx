import type { Device } from "~/lib/model";
import Instructions from "./Instructions";

interface Props {
	device: Device;
}

export default function WindowsInstructions(props: Props) {
	return (
		<>
			<Instructions.Title>
				Configure DNS over HTTPS to enable blkhole.
			</Instructions.Title>
			<Instructions>
				<li>
					Go to{" "}
					<span class="font-medium tracking-wide">
						Settings → Network & Internet → Ethernet/WiFi
					</span>
				</li>
				<li>
					Click{" "}
					<span class="font-medium tracking-wide">
						DNS server assignment → Edit
					</span>
				</li>
				<li>
					Select <span class="font-medium tracking-wide">Manual</span> and
					toggle on <span class="font-medium tracking-wide">IPv4</span>
				</li>
				<li>
					Enter DNS server:{" "}
					<span class="font-medium tracking-wide">
						{`${props.device.hash.toLowerCase()}.blkhole.sh`}
					</span>
				</li>
				<li>
					Toggle <span class="font-medium tracking-wide">DNS over HTTPS</span>{" "}
					to{" "}
					<span class="font-medium tracking-wide">On (automatic template)</span>
				</li>
				<li>
					Click <span class="font-medium tracking-wide">Save</span>
				</li>
			</Instructions>
		</>
	);
}
