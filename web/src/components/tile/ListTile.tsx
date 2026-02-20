import { createSignal } from "solid-js";
import { List } from "~/lib/model";
import ActionButton from "~/components/ui/ActionButton";
import { deleteList } from "~/lib/api";
import DeleteModal from "../modal/DeleteModal";

interface Props {
	list: List;
	onEdit: (list: List) => void;
	onDeleted: () => void;
}

export default function ListTile(props: Props) {
	const [deleteOpen, setDeleteOpen] = createSignal(false);

	return (
		<div class="py-8 flex flex-col gap-6">
			<p class="font-display text-xl">{props.list.name}</p>
			<div class="flex flex-row">
				<div class="flex-1 grid grid-cols-[3fr_1fr] text-zinc-500">
					<p class="pb-2 text-sm tracking-wider">DESCRIPTION</p>
					<p class="pb-2 text-sm tracking-wider">DOMAINS</p>
					<p class="max-w-3xl text-sm">{props.list.description}</p>
					<p class="max-w-xs text-black">{props.list.rules.toLocaleString()}</p>
				</div>
				<div class="flex flex-row gap-6">
					<ActionButton onclick={() => props.onEdit(props.list)}>
						EDIT
					</ActionButton>
					<ActionButton onclick={() => setDeleteOpen(true)}>
						DELETE
					</ActionButton>
				</div>
			</div>
			<DeleteModal
				open={deleteOpen()}
				name={props.list.name}
				onClose={() => setDeleteOpen(false)}
				onConfirm={async () => {
					await deleteList(props.list.id);
					setDeleteOpen(false);
					props.onDeleted();
				}}
			/>
		</div>
	);
}
