import { createResource, createSignal, Show } from "solid-js";
import PageShell from "~/components/layout/PageShell";
import TextInput from "~/components/form/TextInput";
import ButtonSolid from "~/components/ui/ButtonSolid";
import ButtonGhost from "~/components/ui/ButtonGhost";
import Modal from "~/components/modal/Modal";
import Divider from "~/components/ui/Divider";
import { changePassword, deleteAccount, getSettings, logout } from "~/lib/api";
import { useAuth } from "~/context/AuthContext";
import { useNavigate } from "@solidjs/router";

export default function Settings() {
	const auth = useAuth();
	const navigate = useNavigate();

	const [settings] = createResource(getSettings);

	// Password change state
	const [currentPassword, setCurrentPassword] = createSignal("");
	const [newPassword, setNewPassword] = createSignal("");
	const [confirmPassword, setConfirmPassword] = createSignal("");
	const [passwordError, setPasswordError] = createSignal("");
	const [passwordSuccess, setPasswordSuccess] = createSignal(false);
	const [passwordLoading, setPasswordLoading] = createSignal(false);

	// Delete account state
	const [deleteModalOpen, setDeleteModalOpen] = createSignal(false);

	const handlePasswordSubmit = async (e: Event) => {
		e.preventDefault();
		setPasswordError("");
		setPasswordSuccess(false);

		if (newPassword().length < 8) {
			setPasswordError("New password must be at least 8 characters.");
			return;
		}
		if (newPassword() !== confirmPassword()) {
			setPasswordError("New passwords do not match.");
			return;
		}

		setPasswordLoading(true);
		try {
			await changePassword(currentPassword(), newPassword());
			setPasswordSuccess(true);
			setCurrentPassword("");
			setNewPassword("");
			setConfirmPassword("");
		} catch (err: any) {
			const msg: string = err?.message ?? "Failed to change password.";
			// Strip trailing newline that http.Error appends
			setPasswordError(msg.trim());
		} finally {
			setPasswordLoading(false);
		}
	};

	const handleDeleteConfirm = async () => {
		await deleteAccount();
		await logout();
		navigate("/auth/signin", { replace: true });
	};

	const user = () => auth.user();

	return (
		<PageShell title="Settings" description="Fine-tune your singularity.">
			{/* Account section */}
			<section class="py-10 flex flex-col gap-8 max-w-md">
				<h2 class="font-display text-2xl tracking-tight">Account</h2>
				<div class="flex flex-col gap-1">
					<p class="font-medium text-zinc-700 text-sm tracking-wider">Email</p>
					<p class="py-2 text-sm tracking-wider text-zinc-500">
						{user()?.email ?? "—"}
					</p>
					<Divider />
				</div>

				<form onSubmit={handlePasswordSubmit} class="flex flex-col gap-6">
					<p class="font-medium text-zinc-700 text-sm tracking-wider">
						Change Password
					</p>
					<TextInput
						label="Current Password"
						type="password"
						value={currentPassword()}
						onInput={(e) => setCurrentPassword(e.currentTarget.value)}
						placeholder="••••••••"
					/>
					<TextInput
						label="New Password"
						type="password"
						value={newPassword()}
						onInput={(e) => setNewPassword(e.currentTarget.value)}
						placeholder="Min. 8 characters"
						hint="Minimum 8 characters"
					/>
					<TextInput
						label="Confirm New Password"
						type="password"
						value={confirmPassword()}
						onInput={(e) => setConfirmPassword(e.currentTarget.value)}
						placeholder="••••••••"
					/>
					<Show when={passwordError()}>
						<p class="text-sm text-red-700">{passwordError()}</p>
					</Show>
					<Show when={passwordSuccess()}>
						<p class="text-sm text-green-700">Password updated successfully.</p>
					</Show>
					<div class="flex justify-end">
						<ButtonSolid
							type="submit"
							class={`py-2 text-sm${passwordLoading() ? " opacity-50 pointer-events-none" : ""}`}
						>
							{passwordLoading() ? "SAVING..." : "SAVE PASSWORD"}
						</ButtonSolid>
					</div>
				</form>
			</section>

			<Divider />

			{/* DNS Configuration section */}
			<section class="py-10 flex flex-col gap-8 max-w-md">
				<h2 class="font-display text-2xl tracking-tight">DNS Configuration</h2>
				<div class="flex flex-col gap-1">
					<p class="font-medium text-zinc-700 text-sm tracking-wider">
						Upstream DNS Server
					</p>
					<Show
						when={!settings.loading}
						fallback={
							<p class="py-2 text-sm tracking-wider text-zinc-400">Loading…</p>
						}
					>
						<p class="py-2 text-sm tracking-wider font-mono text-zinc-800">
							{settings()?.upstreamDns ?? "—"}
						</p>
					</Show>
					<Divider />
					<p class="text-xs text-zinc-400">
						Queries that are not blocked are forwarded to this server.
					</p>
				</div>
			</section>

			<Divider />

			{/* Danger Zone section */}
			<section class="py-10 flex flex-col gap-8 max-w-md">
				<h2 class="font-display text-2xl tracking-tight text-red-700">
					Danger Zone
				</h2>
				<div class="flex flex-col gap-3">
					<p class="text-sm text-zinc-500">
						Permanently delete your account and all associated data. This cannot
						be undone.
					</p>
					<div>
						<ButtonGhost
							onclick={() => setDeleteModalOpen(true)}
							class="py-2 text-sm text-red-700 px-4 font-medium cursor-pointer"
						>
							DELETE ACCOUNT
						</ButtonGhost>
					</div>
				</div>
			</section>

			<Modal
				title="Delete Account"
				open={deleteModalOpen()}
				onClose={() => setDeleteModalOpen(false)}
				onConfirm={handleDeleteConfirm}
				confirmLabel="DELETE"
			>
				<p class="text-sm text-zinc-500">
					Are you sure you want to permanently delete your account? All your
					devices, lists, and schedules will be removed. This action cannot be
					undone.
				</p>
			</Modal>
		</PageShell>
	);
}
