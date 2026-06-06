import {
	createEffect,
	createResource,
	createSignal,
	For,
	JSX,
	Show,
} from "solid-js";
import MultiStepModal from "./MultiStepModal";
import TextInput from "../form/TextInput";
import TimeInput from "../form/TimeInput";
import CheckboxInput from "../form/CheckboxInput";
import DaysInput, { Days } from "../form/DaysInput";
import {
	createSchedule,
	updateSchedule,
	getDevices,
	getLists,
} from "~/lib/api";
import { required } from "~/lib/validate";
import { Schedule } from "~/lib/model";
import Apple from "../icons/Apple";
import Android from "../icons/Android";
import Linux from "../icons/Linux";
import Windows from "../icons/Windows";

function osIcon(os: string): JSX.Element {
	const lower = os.toLowerCase();
	if (lower.includes("ios") || lower.includes("macos") || lower.includes("mac"))
		return <Apple class="size-5" />;
	if (lower.includes("android")) return <Android class="size-5" />;
	if (lower.includes("linux")) return <Linux class="size-5" />;
	if (lower.includes("windows")) return <Windows class="size-5" />;
	return <></>;
}

interface Props {
	open: boolean;
	schedule?: Schedule | null;
	onClose: () => void;
	onSaved: () => void;
}

const DEFAULT_DAYS = {
	monday: true,
	tuesday: true,
	wednesday: true,
	thursday: true,
	friday: true,
	saturday: false,
	sunday: false,
};

export default function ScheduleModal(props: Props) {
	const [devices] = createResource(getDevices);
	const [lists] = createResource(getLists);
	const [name, setName] = createSignal("");
	const [startTime, setStartTime] = createSignal("08:00");
	const [endTime, setEndTime] = createSignal("17:00");
	const [days, setDays] = createSignal<Days>({ ...DEFAULT_DAYS });
	const [listIds, setListIds] = createSignal<number[]>([]);
	const [deviceIds, setDeviceIds] = createSignal<string[]>([]);
	const [step1Submitted, setStep1Submitted] = createSignal(false);
	const [daysError, setDaysError] = createSignal("");
	const [listsError, setListsError] = createSignal("");
	const [devicesError, setDevicesError] = createSignal("");
	const [error, setError] = createSignal("");

	const isEditMode = () => !!props.schedule;
	const title = () => (isEditMode() ? "Edit Schedule" : "Create Schedule");

	// Pre-fill form when opening in edit mode
	createEffect(() => {
		if (props.schedule && props.open && devices() && lists()) {
			setName(props.schedule.name);
			setStartTime(props.schedule.startTime);
			setEndTime(props.schedule.endTime);
			setDays({
				monday: props.schedule.monday,
				tuesday: props.schedule.tuesday,
				wednesday: props.schedule.wednesday,
				thursday: props.schedule.thursday,
				friday: props.schedule.friday,
				saturday: props.schedule.saturday,
				sunday: props.schedule.sunday,
			});
			// Match list names to list IDs
			const listArray = lists() || [];
			const selectedListIds = listArray
				.filter((l) => props.schedule!.listNames.includes(l.name))
				.map((l) => l.id);
			setListIds(selectedListIds);
			// Match device names to device IDs
			const deviceList = devices() || [];
			const selectedDeviceIds = deviceList
				.filter((d) => props.schedule!.deviceNames.includes(d.name))
				.map((d) => d.id);
			setDeviceIds(selectedDeviceIds);
		} else if (!props.open) {
			// Reset when modal closes
			setName("");
			setStartTime("08:00");
			setEndTime("17:00");
			setDays({ ...DEFAULT_DAYS });
			setListIds([]);
			setDeviceIds([]);
			setStep1Submitted(false);
			setDaysError("");
			setListsError("");
			setDevicesError("");
			setError("");
		}
	});

	const toggleList = (id: number) => {
		setListIds((prev) => {
			const next = prev.includes(id)
				? prev.filter((l) => l !== id)
				: [...prev, id];
			if (next.length > 0) setListsError("");
			return next;
		});
	};

	const toggleDevice = (id: string) => {
		setDeviceIds((prev) => {
			const next = prev.includes(id)
				? prev.filter((d) => d !== id)
				: [...prev, id];
			if (next.length > 0) setDevicesError("");
			return next;
		});
	};

	const handleClose = () => {
		setStep1Submitted(false);
		setDaysError("");
		setListsError("");
		setDevicesError("");
		setName("");
		setStartTime("08:00");
		setEndTime("17:00");
		setDays({ ...DEFAULT_DAYS });
		setListIds([]);
		setDeviceIds([]);
		setError("");
		props.onClose();
	};

	const handleBeforeNext = (step: number) => {
		if (step === 0) {
			setStep1Submitted(true);
			const noDays = !Object.values(days()).some(Boolean);
			if (noDays) setDaysError("Select at least one day");
			else setDaysError("");
			return name().trim() !== "" && !noDays;
		}
		return true;
	};

	const handleSubmit = async () => {
		try {
			setError("");
			if (isEditMode()) {
				await updateSchedule(
					props.schedule!.id,
					name(),
					startTime(),
					endTime(),
					props.schedule!.active,
					days(),
					listIds(),
					deviceIds(),
				);
			} else {
				await createSchedule(
					name(),
					startTime(),
					endTime(),
					days(),
					listIds(),
					deviceIds(),
				);
			}
			setName("");
			setStartTime("08:00");
			setEndTime("17:00");
			setDays({ ...DEFAULT_DAYS });
			setListIds([]);
			setDeviceIds([]);
			setStep1Submitted(false);
			setDaysError("");
			setListsError("");
			props.onSaved();
		} catch {
			setError(`Failed to ${isEditMode() ? "update" : "create"} schedule.`);
		}
	};

	return (
		<MultiStepModal
			title={title()}
			open={props.open}
			onClose={handleClose}
			onConfirm={handleSubmit}
			onBeforeNext={handleBeforeNext}
		>
			{!props.schedule?.isDefault && (
				<MultiStepModal.Step>
					<div class="flex flex-col gap-8">
						<TextInput
							label="SCHEDULE NAME"
							placeholder="School Hours"
							value={name()}
							onInput={(e) => setName(e.currentTarget.value)}
							validate={required()}
							showError={step1Submitted()}
						/>
						<DaysInput
							label="ACTIVE DAYS"
							value={days()}
							onChange={(d) => {
								setDays(d);
								if (Object.values(d).some(Boolean)) setDaysError("");
							}}
							error={daysError()}
						/>
						<div class="flex flex-row gap-8">
							<TimeInput
								class="flex-1"
								label="TIME START"
								value={startTime()}
								onInput={(e) => setStartTime(e.currentTarget.value)}
							/>
							<TimeInput
								class="flex-1"
								label="TIME END"
								value={endTime()}
								onInput={(e) => setEndTime(e.currentTarget.value)}
							/>
						</div>
					</div>
				</MultiStepModal.Step>
			)}
			{lists() && lists()!.length > 0 && (
				<MultiStepModal.Step>
					<div class="flex flex-col gap-1">
						<p class="font-medium text-zinc-700 text-sm tracking-wider">
							SELECT BLOCKLISTS
						</p>
						<div class="flex flex-col divide-y divide-zinc-100">
							<For each={lists()}>
								{(list) => (
									<CheckboxInput
										label={list.name}
										description={list.description}
										checked={listIds().includes(list.id)}
										onChange={() => toggleList(list.id)}
									/>
								)}
							</For>
						</div>
						{listsError() && (
							<p class="text-xs text-red-700 mt-2">{listsError()}</p>
						)}
					</div>
				</MultiStepModal.Step>
			)}
			{devices() && devices()!.length > 0 && (
				<MultiStepModal.Step>
					<div class="flex flex-col gap-1">
						<p class="font-medium text-zinc-700 text-sm tracking-wider">
							APPLY TO DEVICES
						</p>
						<div class="flex flex-col divide-y divide-zinc-100">
							<For each={devices()}>
								{(device) => (
									<CheckboxInput
										label={device.name}
										icon={osIcon(device.os)}
										checked={deviceIds().includes(device.id)}
										onChange={() => toggleDevice(device.id)}
									/>
								)}
							</For>
						</div>
						{devicesError() && (
							<p class="text-xs text-red-700 mt-2">{devicesError()}</p>
						)}
						{error() && <p class="text-xs text-red-700 mt-2">{error()}</p>}
					</div>
				</MultiStepModal.Step>
			)}
		</MultiStepModal>
	);
}
