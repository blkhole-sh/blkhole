import { For, Show } from "solid-js";
import { Schedule } from "~/lib/model";
import ActionButton from "~/components/ActionButton";
import Tag from "~/components/Tag";
import Switch from "~/components/Switch";

interface Props {
	schedule: Schedule;
}

const days = [
	{ label: "M", key: "monday" },
	{ label: "T", key: "tuesday" },
	{ label: "W", key: "wednesday" },
	{ label: "T", key: "thursday" },
	{ label: "F", key: "friday" },
	{ label: "S", key: "saturday" },
	{ label: "S", key: "sunday" },
] as const;

const formatTime = (t: string) => {
	const [hours, minutes] = t.split(":");
	const d = new Date();
	d.setHours(Number(hours), Number(minutes), 0, 0);
	return d.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
};

export default function ScheduleTile(props: Props) {
	return (
		<div class="py-8 flex flex-col gap-8 font-inter">
			<div class="flex flex-col gap-1">
				<div class="flex flex-row justify-between items-center">
					<p class="font-hedvig text-xl">{props.schedule.name}</p>
					<Switch />
				</div>
				<p class="text-zinc-500 text-sm">Active now</p>
			</div>
			<div class="flex flex-row">
				<div class="flex-1 grid grid-cols-3 items-start text-zinc-500">
					<p class="pb-2 text-sm tracking-wider">TIMING</p>
					<p class="pb-2 text-sm tracking-wider">DEVICES</p>
					<p class="pb-2 text-sm tracking-wider">BLOCKLISTS</p>
					<div class="max-w-md">
						<p class="pb-4 text-black">
							{formatTime(props.schedule.startTime)} –{" "}
							{formatTime(props.schedule.endTime)}
						</p>
						<div class="flex flex-row gap-3 text-zinc-300">
							<For each={days}>
								{({ label, key }) => (
									<span classList={{ "text-black": props.schedule[key] }}>
										{label}
									</span>
								)}
							</For>
						</div>
					</div>
					<div class="flex flex-row flex-wrap gap-2 max-w-sm">
						<Show
							when={props.schedule.deviceNames.length > 0}
							fallback={<p>—</p>}
						>
							<For each={props.schedule.deviceNames}>
								{(name) => <Tag>{name}</Tag>}
							</For>
						</Show>
					</div>
					<div class="flex flex-row flex-wrap gap-2 max-w-sm">
						<Show
							when={props.schedule.listNames.length > 0}
							fallback={<p>—</p>}
						>
							<For each={props.schedule.listNames}>
								{(name) => <Tag>{name}</Tag>}
							</For>
						</Show>
					</div>
				</div>
				<div class="flex flex-row gap-6">
					<ActionButton>EDIT</ActionButton>
					<ActionButton>DELETE</ActionButton>
				</div>
			</div>
		</div>
	);
}
