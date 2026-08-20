import { createResource, createSignal, For, Show } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import ScheduleModal from "~/components/modal/ScheduleModal";
import ScheduleTile from "~/components/tile/ScheduleTile";
import EmptyState from "~/components/ui/EmptyState";
import { getSchedules } from "~/lib/api";
import { useScrollToHash } from "~/lib/hooks";
import type { Schedule } from "~/lib/model";

export default function Schedules() {
	useScrollToHash();
	const [modalOpen, setModalOpen] = createSignal(false);
	const [editingSchedule, setEditingSchedule] = createSignal<Schedule | null>(
		null,
	);
	const [schedules, { refetch }] = createResource(getSchedules);

	const handleCreate = () => {
		setEditingSchedule(null);
		setModalOpen(true);
	};

	const handleEdit = (schedule: Schedule) => {
		setEditingSchedule(schedule);
		setModalOpen(true);
	};

	const handleSaved = () => {
		setModalOpen(false);
		setEditingSchedule(null);
		refetch();
	};

	return (
		<PageShell
			title="Schedules"
			description="Control when your black hole is hungry."
			cta="CREATE SCHEDULE"
			onCTA={handleCreate}
		>
			<Show
				when={(schedules()?.length ?? 0) > 0}
				fallback={
					<Show when={!schedules.loading}>
						<EmptyState
							message="blkhole exists outside of time"
							subtitle="Create a schedule to control when blocking occurs"
						/>
					</Show>
				}
			>
				<div class="-mt-8 divide-y divide-zinc-200 border-b border-zinc-200">
					<For each={schedules()}>
						{(schedule) => (
							<ScheduleTile
								schedule={schedule}
								onEdit={handleEdit}
								onDeleted={refetch}
								onUpdated={refetch}
							/>
						)}
					</For>
				</div>
			</Show>
			<ScheduleModal
				open={modalOpen()}
				schedule={editingSchedule()}
				onClose={() => setModalOpen(false)}
				onSaved={handleSaved}
			/>
		</PageShell>
	);
}
