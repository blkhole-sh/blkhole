import { createSignal } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import DeleteAccountModal from "~/components/modal/DeleteAccountModal";
import AccountSettings from "~/components/settings/AccountSettings";
import DangerZone from "~/components/settings/DangerZone";
import DNSSettings from "~/components/settings/DNSSettings";

export default function Settings() {
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);

	return (
		<PageShell title="Settings" description="Fine-tune your singularity.">
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-12">
				<AccountSettings />
				<div class="flex flex-col gap-12">
					<DNSSettings />
					<DangerZone onDelete={() => setDeleteModalOpen(true)} />
				</div>
			</div>
			<DeleteAccountModal
				open={deleteModalOpen()}
				onClose={() => setDeleteModalOpen(false)}
			/>
		</PageShell>
	);
}
