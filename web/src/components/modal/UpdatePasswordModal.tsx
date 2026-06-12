import { createSignal, Show } from "solid-js";
import Modal from "./Modal";
import TextInput from "../form/TextInput";
import { compose, minLength, required } from "~/lib/validate";
import { changePassword } from "~/lib/api";

interface Props {
	open: boolean;
	onClose: () => void;
}

export default function UpdatePasswordModal(props: Props) {
	const [currentPassword, setCurrentPassword] = createSignal("");
	const [newPassword, setNewPassword] = createSignal("");
	const [confirmPassword, setConfirmPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal<string | undefined>(undefined);
	const [loading, setLoading] = createSignal(false);

	const passwordsMatch = () =>
		newPassword() === confirmPassword() ? undefined : "Passwords do not match";

	const isInvalid = () =>
		!currentPassword() ||
		!!compose(required(), minLength(12))(newPassword()) ||
		!!passwordsMatch();

	const resetForm = () => {
		setCurrentPassword("");
		setNewPassword("");
		setConfirmPassword("");
		setSubmitted(false);
		setError(undefined);
	};

	const handleConfirm = async () => {
		if (loading()) return;
		setSubmitted(true);
		setError(undefined);

		if (isInvalid()) return;

		setLoading(true);
		try {
			await changePassword(currentPassword(), newPassword());
			resetForm();
			props.onClose();
		} catch (err) {
			const msg = err instanceof Error ? err.message : String(err);
			setError(
				msg.includes("incorrect") || msg.includes("400")
					? "Current password is incorrect."
					: "Failed to change password. Please try again.",
			);
		} finally {
			setLoading(false);
		}
	};

	return (
		<Modal
			title="Change Password"
			open={props.open}
			onClose={() => {
				resetForm();
				props.onClose();
			}}
			onConfirm={handleConfirm}
			confirmLabel={loading() ? "UPDATING..." : "UPDATE PASSWORD"}
		>
			<div class="flex flex-col gap-4">
				<TextInput
					label="CURRENT PASSWORD"
					type="password"
					placeholder="••••••••"
					value={currentPassword()}
					onInput={(e) => setCurrentPassword(e.currentTarget.value)}
					validate={required()}
					showError={submitted()}
				/>
				<TextInput
					label="NEW PASSWORD"
					type="password"
					placeholder="••••••••"
					value={newPassword()}
					onInput={(e) => setNewPassword(e.currentTarget.value)}
					validate={compose(required(), minLength(12))}
					showError={submitted()}
				/>
				<TextInput
					label="CONFIRM NEW PASSWORD"
					type="password"
					placeholder="••••••••"
					value={confirmPassword()}
					onInput={(e) => setConfirmPassword(e.currentTarget.value)}
					validate={compose(required(), () => passwordsMatch())}
					showError={submitted()}
				/>
				<Show when={error()}>
					<p class="text-sm text-red-700">{error()}</p>
				</Show>
			</div>
		</Modal>
	);
}
