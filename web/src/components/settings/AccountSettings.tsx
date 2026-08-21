import { createSignal } from "solid-js";
import TextInput from "../form/TextInput";
import UpdatePasswordModal from "../modal/UpdatePasswordModal";
import ActionButton from "../ui/ActionButton";

interface Props {
	email: string;
	onEmailInput: (value: string) => void;
	emailError?: string | undefined;
	onDelete: () => void;
}

export default function AccountSettings(props: Props) {
	const [passwordModalOpen, setPasswordModalOpen] = createSignal(false);

	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-medium text-sm tracking-wider text-zinc-700">ACCOUNT</h2>
			<TextInput
				label="EMAIL"
				type="email"
				value={props.email}
				onInput={(event) => props.onEmailInput(event.currentTarget.value)}
				error={props.emailError}
			/>
			<div class="flex flex-row items-center gap-6">
				<ActionButton onclick={() => setPasswordModalOpen(true)}>
					CHANGE PASSWORD
				</ActionButton>
				<ActionButton onclick={props.onDelete}>DELETE ACCOUNT</ActionButton>
			</div>
			<UpdatePasswordModal
				open={passwordModalOpen()}
				onClose={() => setPasswordModalOpen(false)}
			/>
		</section>
	);
}
