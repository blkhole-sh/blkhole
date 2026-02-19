import { List } from "~/lib/model";
import ActionButton from "~/components/ActionButton";

interface Props {
	list: List;
}

export default function ListTile(props: Props) {
	return (
		<div class="py-8 flex flex-col gap-6 font-inter">
			<p class="font-hedvig text-xl">{props.list.name}</p>
			<div class="flex flex-row">
				<div class="flex-1 grid grid-cols-[3fr_1fr] text-zinc-500">
					<p class="pb-2 text-sm tracking-wider">DESCRIPTION</p>
					<p class="pb-2 text-sm tracking-wider">DOMAINS</p>
					<p class="max-w-3xl text-sm">{props.list.description}</p>
					<p class="max-w-xs text-black">{props.list.rules.toLocaleString()}</p>
				</div>
				<div class="flex flex-row gap-6">
					<ActionButton>EDIT</ActionButton>
					<ActionButton>DELETE</ActionButton>
				</div>
			</div>
		</div>
	);
}
