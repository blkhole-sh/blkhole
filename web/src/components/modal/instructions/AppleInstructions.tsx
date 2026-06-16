import { Show } from "solid-js";
import type { Device } from "~/lib/model";
import { configUrl, isIOS, isMacOS } from "~/lib/utils";
import Instructions from "./Instructions";

interface Props {
	device: Device;
}

export default function AppleInstructions(props: Props) {
	return (
		<>
			<Instructions.Title>
				Download and install the configuration profile to enable blkhole.
			</Instructions.Title>
			<a
				href={configUrl(props.device.id)}
				class="text-sm text-black font-medium underline tracking-wider hover:text-zinc-600"
				download=""
			>
				CONFIGURATION PROFILE
			</a>
			<Instructions>
				<Show when={isIOS(props.device.os)}>
					<li>Tap the downloaded .mobileconfig file</li>
					<li>
						Go to{" "}
						<span class="font-medium tracking-wide">
							Settings → Profile Downloaded
						</span>
					</li>
					<li>
						Tap <span class="font-medium tracking-wide">Install</span> and
						follow the prompts
					</li>
				</Show>
				<Show when={isMacOS(props.device.os)}>
					<li>Open the downloaded .mobileconfig file</li>
					<li>
						Go to{" "}
						<span class="font-medium tracking-wide">
							System Settings → Privacy & Security → Profiles
						</span>
					</li>
					<li>
						Select the profile and click{" "}
						<span class="font-medium tracking-wide">Install</span>
					</li>
				</Show>
			</Instructions>
		</>
	);
}
