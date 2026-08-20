import { createSignal } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import DeleteAccountModal from "~/components/modal/DeleteAccountModal";
import AccountSettings from "~/components/settings/AccountSettings";
import DNSSettings from "~/components/settings/DNSSettings";

export default function Settings() {
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);
	const version = __APP_VERSION__;
	const revisionMeta = document
		.querySelector('meta[name="blkhole-revision"]')
		?.getAttribute("content");
	const revision =
		revisionMeta && revisionMeta !== "__BLKHOLE_REVISION__"
			? revisionMeta.toUpperCase()
			: null;

	return (
		<PageShell title="Settings" description="Fine-tune your singularity.">
			<div class="flex flex-col flex-1 min-h-0">
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-12 max-w-6xl">
					<AccountSettings onDelete={() => setDeleteModalOpen(true)} />
					<DNSSettings />
				</div>
				<div class="mt-auto -mb-4 pt-8 flex flex-row justify-between border-t border-zinc-200 text-sm tracking-wider text-zinc-400">
					<span>{version ? `VERSION ${version}` : ""}</span>
					<span>{revision}</span>
				</div>
			</div>
			<DeleteAccountModal
				open={deleteModalOpen()}
				onClose={() => setDeleteModalOpen(false)}
			/>
		</PageShell>
	);
}
