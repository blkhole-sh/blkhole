import { createSignal } from "solid-js";
import { useAuth } from "~/context/AuthContext";
import UpdatePasswordModal from "../modal/UpdatePasswordModal";
import ActionButton from "../ui/ActionButton";
import Divider from "../ui/Divider";

interface Props {
	onDelete: () => void;
}

export default function AccountSettings(props: Props) {
	const { user } = useAuth();
	const [passwordModalOpen, setPasswordModalOpen] = createSignal(false);

	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-medium tracking-wider">ACCOUNT</h2>
			<div class="flex flex-col gap-1">
				<p class="font-medium text-zinc-700 text-sm tracking-wider">EMAIL</p>
				<p class="py-2 text-sm leading-snug tracking-wider text-zinc-500">
					{user()?.email}
				</p>
				<Divider />
			</div>
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
