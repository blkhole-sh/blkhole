import { createEffect, createResource, createSignal, Show } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import DeleteAccountModal from "~/components/modal/DeleteAccountModal";
import AccountSettings from "~/components/settings/AccountSettings";
import DNSSettings from "~/components/settings/DNSSettings";
import { useAuth } from "~/context/AuthContext";
import { getSettings, updateEmail, updateSettings } from "~/lib/api";
import { compose, required, email as validEmail } from "~/lib/validate";

export default function Settings() {
	const auth = useAuth();
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);
	const [email, setEmail] = createSignal(auth.user()?.email ?? "");
	const [upstreamDNS, setUpstreamDNS] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [saving, setSaving] = createSignal(false);
	const [saveError, setSaveError] = createSignal<string>();
	const [settings] = createResource(getSettings);
	const version = __APP_VERSION__;
	const revisionMeta = document
		.querySelector('meta[name="blkhole-revision"]')
		?.getAttribute("content");
	const revision =
		revisionMeta && revisionMeta !== "__BLKHOLE_REVISION__"
			? revisionMeta.toUpperCase()
			: null;
	const emailError = () => compose(required(), validEmail())(email().trim());
	const upstreamDNSError = () => {
		const value = upstreamDNS().trim();
		if (!value) return "Required";
		if (
			!(
				/^(?:\d{1,3}\.){3}\d{1,3}:\d+$/.test(value) ||
				/^\[[0-9a-f:]+\]:\d+$/i.test(value)
			)
		)
			return "Use IP:port";
	};

	createEffect(() => {
		const value = settings()?.upstreamDns;
		if (value) setUpstreamDNS(value);
	});

	const handleSave = async () => {
		if (saving()) return;
		setSubmitted(true);
		setSaveError(undefined);
		if (emailError() || upstreamDNSError()) return;

		setSaving(true);
		try {
			const [user] = await Promise.all([
				updateEmail(email().trim()),
				updateSettings(upstreamDNS().trim()),
			]);
			auth.updateUser(user);
			setSubmitted(false);
		} catch {
			setSaveError("Unable to save settings. Please try again.");
		} finally {
			setSaving(false);
		}
	};

	return (
		<PageShell
			title="Settings"
			description="Fine-tune your singularity."
			cta={saving() ? "SAVING..." : "SAVE"}
			onCTA={handleSave}
		>
			<div class="flex flex-col flex-1 min-h-0">
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-12 max-w-6xl">
					<AccountSettings
						email={email()}
						onEmailInput={setEmail}
						emailError={submitted() ? emailError() : undefined}
						onDelete={() => setDeleteModalOpen(true)}
					/>
					<DNSSettings
						upstreamDNS={upstreamDNS()}
						onUpstreamDNSInput={setUpstreamDNS}
						error={submitted() ? upstreamDNSError() : undefined}
					/>
				</div>
				<Show when={saveError()}>
					<p class="pt-6 text-sm text-red-700">{saveError()}</p>
				</Show>
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
