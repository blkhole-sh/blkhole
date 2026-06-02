import { createSignal } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import DNSSettings from "~/components/settings/DNSSettings";
import DangerZone from "~/components/settings/DangerZone";
import AccountSettings from "~/components/settings/AccountSettings";
import DeleteAccountModal from "~/components/modal/DeleteAccountModal";

export default function Settings() {
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);

	return (
		<PageShell title="Settings" description="Fine-tune your singularity.">
			<div class="py-8 grid grid-cols-1 lg:grid-cols-2 gap-12">
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
