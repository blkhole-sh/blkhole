import { A } from "@solidjs/router";
import Dot from "../icons/Dot";
import Edit from "../icons/Edit";
import Trash from "../icons/Trash";
import { Schedule } from "~/lib/model";

interface Props {
	schedule: Schedule;
}

export default function ScheduleTile(props: Props) {
	const time = () => `${props.schedule.startTime} - ${props.schedule.endTime}`;

	const activeDays = () => {
		const daysMap = {
			monday: "Mon",
			tuesday: "Tue",
			wednesday: "Wed",
			thursday: "Thu",
			friday: "Fri",
			saturday: "Sat",
			sunday: "Sun",
		};

		const days = Object.entries(daysMap)
			.filter(([dayKey]) => props.schedule[dayKey as keyof Schedule])
			.map(([_, abbreviation]) => abbreviation);

		return days.join(", ");
	};

	const listsAndDomains = () => {
		const lists = props.schedule.lists;
		const domains = props.schedule.domains;

		if (lists === 0 && domains === 0) {
			return "None"; // Optional: handle case when both are empty
		}

		if (lists === 0) {
			return `${domains} domains`;
		}

		if (domains === 0) {
			return `${lists} lists`;
		}

		return `${lists} lists, ${domains} domains`;
	};

	return (
		<li class="flex flex-wrap items-center justify-between gap-x-6 gap-y-4 py-2 sm:flex-nowrap">
			<div class="flex-1">
				<p class="text-sm/6 font-semibold text-gray-900">
					{props.schedule.name}
				</p>
				<div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
					<p>{time()}</p>
					<Dot />
					<p>{activeDays()}</p>
					<Dot />
					<p>{listsAndDomains()}</p>
				</div>
			</div>
			<div class="flex gap-x-4">
				<A href={`/schedules/${props.schedule.id}`}>
					<Edit />
				</A>
				<Trash />
			</div>
		</li>
	);
}
