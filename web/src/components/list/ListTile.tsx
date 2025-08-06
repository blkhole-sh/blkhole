import { Show } from "solid-js";
import { List } from "~/lib/model";
import Dot from "../icons/Dot";
import Edit from "../icons/Edit";
import Trash from "../icons/Trash";
import { A } from "@solidjs/router";

interface Props {
	list: List;
}

export default function ListTile(props: Props) {
	const domains = () => {
		if (props.list.rules === 0) {
			return "Empty";
		}

		return `${props.list.rules} rules`;
	};

	return (
		<li class="flex flex-wrap items-center justify-between gap-x-6 gap-y-4 py-2 sm:flex-nowrap">
			<div class="flex-1">
				<p class="text-sm/6 font-semibold text-gray-900">{props.list.name}</p>
				<p class="text-sm/6 font-normal text-gray-500">
					{props.list.description}
				</p>
				<div class="mt-1 flex items-center gap-x-2 text-xs/5 text-gray-500">
					<Show when={props.list.source}>
						{(source) => (
							<>
								<A
									href={source()}
									class="text-blue-800 hover:text-blue-900 underline"
								>
									{source()}
								</A>
								<Dot />
							</>
						)}
					</Show>
					<p>{domains()}</p>
				</div>
			</div>
			<div class="flex gap-x-4">
				<A href={`/lists/${props.list.id}`}>
					<Edit />
				</A>
				<Trash />
			</div>
		</li>
	);
}
