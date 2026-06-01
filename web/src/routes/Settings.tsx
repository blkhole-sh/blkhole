import { createResource, createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import PageShell from "~/components/layout/PageShell";
import TextInput from "~/components/form/TextInput";
import ButtonSolid from "~/components/ui/ButtonSolid";
import Modal from "~/components/modal/Modal";
import Divider from "~/components/ui/Divider";
import { getSettings, changePassword, deleteAccount, logout } from "~/lib/api";
import { compose, required, minLength } from "~/lib/validate";

function getCurrentUserEmail(): string {
	try {
		const raw = localStorage.getItem("user");
		if (!raw) return "";
		return JSON.parse(raw).email ?? "";
	} catch {
		return "";
	}
}

export default function Settings() {
	const navigate = useNavigate();

	// DNS settings resource
	const [settings] = createResource(getSettings);

	// Password change form
	const [currentPassword, setCurrentPassword] = createSignal("");
	const [newPassword, setNewPassword] = createSignal("");
	const [confirmPassword, setConfirmPassword] = createSignal("");
	const [passwordSubmitted, setPasswordSubmitted] = createSignal(false);
	const [passwordError, setPasswordError] = createSignal<string | undefined>(undefined);
	const [passwordSuccess, setPasswordSuccess] = createSignal(false);
	const [passwordLoading, setPasswordLoading] = createSignal(false);

	// Delete account modal
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);

	const passwordsMatch = () =>
		newPassword() === confirmPassword() ? undefined : "Passwords do not match";

	const handleChangePassword = async (e: Event) => {
		e.preventDefault();
		setPasswordSubmitted(true);
		setPasswordError(undefined);
		setPasswordSuccess(false);

		if (
			!currentPassword() ||
			compose(required(), minLength(8))(newPassword()) ||
			passwordsMatch()
		) {
			return;
		}

		setPasswordLoading(true);
		try {
			await changePassword(currentPassword(), newPassword());
			setPasswordSuccess(true);
			setCurrentPassword("");
			setNewPassword("");
			setConfirmPassword("");
			setPasswordSubmitted(false);
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			if (msg.includes("incorrect") || msg.includes("400")) {
				setPasswordError("Current password is incorrect.");
			} else {
				setPasswordError("Failed to change password. Please try again.");
			}
		} finally {
			setPasswordLoading(false);
		}
	};

	const handleDeleteAccount = async () => {
		await deleteAccount();
		await logout();
		navigate("/auth/signin", { replace: true });
	};

	return (
		<PageShell title="Settings" description="Fine-tune your singularity.">
			<div class="py-8 flex flex-col gap-12 max-w-xl">
				{/* Account section */}
				<section class="flex flex-col gap-6">
					<h2 class="font-display text-2xl">Account</h2>
					<div class="flex flex-col gap-1">
						<p class="font-medium text-zinc-700 text-sm tracking-wider">EMAIL</p>
						<p class="py-2 text-sm leading-snug tracking-wider text-zinc-500">
							{getCurrentUserEmail()}
						</p>
						<Divider />
					</div>

					<form onSubmit={handleChangePassword} class="flex flex-col gap-4">
						<p class="font-medium text-zinc-700 text-sm tracking-wider">CHANGE PASSWORD</p>
						<TextInput
							label="CURRENT PASSWORD"
							type="password"
							placeholder="••••••••"
							value={currentPassword()}
							onInput={(e) => setCurrentPassword(e.currentTarget.value)}
							validate={required()}
							showError={passwordSubmitted()}
						/>
						<TextInput
							label="NEW PASSWORD"
							type="password"
							placeholder="••••••••"
							value={newPassword()}
							onInput={(e) => setNewPassword(e.currentTarget.value)}
							validate={compose(required(), minLength(8))}
							showError={passwordSubmitted()}
						/>
						<TextInput
							label="CONFIRM NEW PASSWORD"
							type="password"
							placeholder="••••••••"
							value={confirmPassword()}
							onInput={(e) => setConfirmPassword(e.currentTarget.value)}
							validate={compose(required(), () => passwordsMatch())}
							showError={passwordSubmitted()}
						/>
						<Show when={passwordError()}>
							<p class="text-sm text-red-700">{passwordError()}</p>
						</Show>
						<Show when={passwordSuccess()}>
							<p class="text-sm text-zinc-500">Password updated successfully.</p>
						</Show>
						<div>
							<ButtonSolid
								type="submit"
								class={
									passwordLoading()
										? "opacity-50 pointer-events-none py-2 text-sm"
										: "py-2 text-sm"
								}
							>
								UPDATE PASSWORD
							</ButtonSolid>
						</div>
					</form>
				</section>

				<Divider />

				{/* DNS Configuration section */}
				<section class="flex flex-col gap-6">
					<h2 class="font-display text-2xl">DNS Configuration</h2>
					<div class="flex flex-col gap-1">
						<p class="font-medium text-zinc-700 text-sm tracking-wider">
							UPSTREAM DNS SERVER
						</p>
						<p class="py-2 text-sm leading-snug tracking-wider text-zinc-500">
							<Show when={!settings.loading} fallback="Loading...">
								{settings()?.upstreamDns ?? "—"}
							</Show>
						</p>
						<Divider />
						<p class="text-xs text-zinc-400">
							DNS queries not matched by your blocklists are forwarded to this server.
						</p>
					</div>
				</section>

				<Divider />

				{/* Danger Zone section */}
				<section class="flex flex-col gap-6">
					<h2 class="font-display text-2xl text-red-600">Danger Zone</h2>
					<div class="flex flex-col gap-2">
						<p class="text-sm text-zinc-500">
							Permanently delete your account and all associated data. This cannot be
							undone.
						</p>
						<div>
							<ButtonSolid
								class="py-2 text-sm bg-red-600"
								onclick={() => setDeleteModalOpen(true)}
							>
								DELETE ACCOUNT
							</ButtonSolid>
						</div>
					</div>
				</section>
			</div>

			<Modal
				title="Delete Account"
				open={deleteModalOpen()}
				onClose={() => setDeleteModalOpen(false)}
				onConfirm={handleDeleteAccount}
				confirmLabel="DELETE"
			>
				<p class="text-sm text-zinc-500">
					Are you sure you want to delete your account? All your devices, lists, and
					schedules will be permanently removed. This action cannot be undone.
				</p>
			</Modal>
		</PageShell>
	);
}
