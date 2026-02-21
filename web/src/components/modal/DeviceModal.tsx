import { createEffect, createSignal } from "solid-js";
import Modal from "./Modal";
import TextInput from "../form/TextInput";
import TileInput, { TileOption } from "../form/TileInput";
import Android from "../icons/Android";
import Linux from "../icons/Linux";
import Windows from "../icons/Windows";
import { createDevice, updateDevice } from "~/lib/api";
import { required } from "~/lib/validate";
import { Device } from "~/lib/model";
import Apple from "../icons/Apple";

interface Props {
	open: boolean;
	device?: Device | null;
	onClose: () => void;
	onSaved: () => void;
}

const OS_OPTIONS: TileOption[] = [
	{ value: "macOS", label: "macOS", icon: () => <Apple /> },
	{ value: "iOS", label: "iOS", icon: () => <Apple /> },
	{ value: "Android", label: "Android", icon: () => <Android /> },
	{ value: "Linux", label: "Linux", icon: () => <Linux /> },
	{ value: "Windows", label: "Windows", icon: () => <Windows /> },
];

export default function DeviceModal(props: Props) {
	const [name, setName] = createSignal("");
	const [os, setOs] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [osError, setOsError] = createSignal("");
	const [error, setError] = createSignal("");

	const isEditMode = () => !!props.device;
	const title = () => (isEditMode() ? "Edit Device" : "Add Device");

	// Pre-fill form when opening in edit mode
	createEffect(() => {
		if (props.device && props.open) {
			setName(props.device.name);
			setOs(props.device.os);
		} else if (!props.open) {
			// Reset when modal closes
			setName("");
			setOs("");
			setSubmitted(false);
			setOsError("");
			setError("");
		}
	});

	const handleClose = () => {
		setSubmitted(false);
		setOsError("");
		setError("");
		setName("");
		setOs("");
		props.onClose();
	};

	const handleOsChange = (value: string) => {
		setOs(value);
		setOsError("");
	};

	const handleSubmit = async () => {
		setSubmitted(true);
		const nameValid = name().trim() !== "";
		const osValid = os() !== "";
		if (!osValid) setOsError("Required");
		if (!nameValid || !osValid) return;

		try {
			setError("");
			if (isEditMode()) {
				await updateDevice(props.device!.id, name(), os());
			} else {
				await createDevice(name(), os());
			}
			setName("");
			setOs("");
			setSubmitted(false);
			setOsError("");
			props.onSaved();
		} catch {
			setError(`Failed to ${isEditMode() ? "update" : "create"} device.`);
		}
	};

	return (
		<Modal
			title={title()}
			open={props.open}
			onClose={handleClose}
			onConfirm={handleSubmit}
		>
			<div class="flex flex-col gap-8">
				<TextInput
					label="DEVICE NAME"
					placeholder="My MacBook"
					value={name()}
					onInput={(e) => setName(e.currentTarget.value)}
					validate={required()}
					showError={submitted()}
				/>
				<TileInput
					label="OPERATING SYSTEM"
					options={OS_OPTIONS}
					value={os()}
					onChange={handleOsChange}
					error={osError()}
				/>
				{error() && <p class="text-sm text-red-700 mt-2">{error()}</p>}
			</div>
		</Modal>
	);
}
