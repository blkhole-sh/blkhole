import { createResource, createSignal, For, Show } from "solid-js";
import ListTile from "~/components/tile/ListTile";
import PageShell from "~/components/layout/PageShell";
import ListModal from "~/components/modal/ListModal";
import EmptyState from "~/components/ui/EmptyState";
import { getLists } from "~/lib/api";
import { List } from "~/lib/model";

export default function Lists() {
	const [modalOpen, setModalOpen] = createSignal(false);
	const [editingList, setEditingList] = createSignal<List | null>(null);
	const [lists, { refetch }] = createResource(getLists);

	const handleCreate = () => {
		setEditingList(null);
		setModalOpen(true);
	};

	const handleEdit = (list: List) => {
		setEditingList(list);
		setModalOpen(true);
	};

	const handleSaved = () => {
		setModalOpen(false);
		setEditingList(null);
		refetch();
	};

	return (
		<PageShell
			title="Blocklists"
			description="Domains that get pulled into the void."
			cta="CREATE BLOCKLIST"
			onCTA={handleCreate}
		>
			<Show
				when={lists()?.length > 0}
				fallback={
					<EmptyState
						message="blkhole is letting everything escape"
						subtitle="Create your first blocklist to start filtering domains"
					/>
				}
			>
				<div class="divide-y divide-zinc-200">
					<For each={lists()}>
						{(list) => (
							<ListTile list={list} onEdit={handleEdit} onDeleted={refetch} />
						)}
					</For>
				</div>
			</Show>
			<ListModal
				open={modalOpen()}
				list={editingList()}
				onClose={() => setModalOpen(false)}
				onSaved={handleSaved}
			/>
		</PageShell>
	);
}
