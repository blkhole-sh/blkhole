import { Match, Switch } from "solid-js";
import type { Device } from "~/lib/model";
import { isAndroid, isApple, isLinux, isWindows } from "~/lib/utils";
import BrowserExtensionSetup from "./BrowserExtensionSetup";
import AndroidInstructions from "./instructions/AndroidInstructions";
import AppleInstructions from "./instructions/AppleInstructions";
import LinuxInstructions from "./instructions/LinuxInstructions";
import WindowsInstructions from "./instructions/WindowsInstructions";
import Modal from "./Modal";

interface Props {
	open: boolean;
	device: Device;
	onClose: () => void;
}

export default function DeviceSetupModal(props: Props) {
	const title = () => {
		if (isApple(props.device.os)) return "Apple Setup";
		if (isAndroid(props.device.os)) return "Android Setup";
		if (isLinux(props.device.os)) return "Linux Setup";
		if (isWindows(props.device.os)) return "Windows Setup";
		return "Device Setup";
	};

	return (
		<Modal
			title={title()}
			open={props.open}
			onClose={props.onClose}
			onConfirm={props.onClose}
			confirmLabel="DONE"
		>
			<div class="flex flex-col gap-8">
				<Switch
					fallback={
						<p class="text-sm text-zinc-500">
							No setup instructions available yet for this device.
						</p>
					}
				>
					<Match when={isApple(props.device.os)}>
						<AppleInstructions device={props.device} />
					</Match>
					<Match when={isAndroid(props.device.os)}>
						<AndroidInstructions device={props.device} />
					</Match>
					<Match when={isLinux(props.device.os)}>
						<LinuxInstructions device={props.device} />
					</Match>
					<Match when={isWindows(props.device.os)}>
						<WindowsInstructions device={props.device} />
					</Match>
				</Switch>
				<BrowserExtensionSetup device={props.device} open={props.open} />
			</div>
		</Modal>
	);
}
