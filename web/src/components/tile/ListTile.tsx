import { createSignal, Show } from "solid-js";
import ActionButton from "~/components/ui/ActionButton";
import Tag from "~/components/ui/Tag";
import { deleteList } from "~/lib/api";
import { useScrollToHash } from "~/lib/hooks";
import type { List } from "~/lib/model";
import DeleteModal from "../modal/DeleteModal";

interface Props {
	list: List;
	onEdit: (list: List) => void;
	onDeleted: () => void;
}

export default function ListTile(props: Props) {
	useScrollToHash();
	const [deleteOpen, setDeleteOpen] = createSignal(false);

	return (
		<div
			id={`list-${props.list.id}`}
			class="py-5 flex flex-row items-center gap-8"
		>
			<div class="flex-1 min-w-0 flex flex-col gap-1">
				<div class="min-h-6.5 flex flex-row items-center gap-3">
					<p class="font-medium tracking-wider truncate">{props.list.name}</p>
					<Show when={props.list.isDefault}>
						<Tag>DEFAULT</Tag>
					</Show>
				</div>
				<Show when={props.list.isDefault}>
					<p class="max-w-xl text-sm leading-normal text-zinc-500">
						{props.list.description}
					</p>
				</Show>
			</div>
			<p class="w-35 flex-shrink-0 tracking-wider">
				{props.list.rules.toLocaleString()} rules
			</p>
			<div class="w-35 flex-shrink-0 flex flex-row justify-end gap-6">
				<Show when={!props.list.isDefault}>
					<ActionButton onclick={() => props.onEdit(props.list)}>
						EDIT
					</ActionButton>
					<ActionButton onclick={() => setDeleteOpen(true)}>
						DELETE
					</ActionButton>
				</Show>
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
