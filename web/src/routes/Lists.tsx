import { createResource, createSignal, For, Show } from "solid-js";
import ListTile from "~/components/tile/ListTile";
import PageShell from "~/components/layout/PageShell";
import ListModal from "~/components/modal/ListModal";
import EmptyState from "~/components/ui/EmptyState";
import { getLists } from "~/lib/api";
import { List } from "~/lib/model";
import { useScrollToHash } from "~/lib/hooks";

export default function Lists() {
	useScrollToHash();
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
				when={lists() && lists()!.length > 0}
				fallback={
					<Show when={!lists.loading}>
						<EmptyState
							message="blkhole is letting everything escape"
							subtitle="Create a blocklist to start pulling domains into the void"
						/>
					</Show>
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
