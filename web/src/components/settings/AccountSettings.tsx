import { createSignal } from "solid-js";
import Divider from "../ui/Divider";
import { useAuth } from "~/context/AuthContext";
import UpdatePasswordModal from "../modal/UpdatePasswordModal";

export default function AccountSettings() {
	const { user } = useAuth();
	const [passwordModalOpen, setPasswordModalOpen] = createSignal(false);

	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-medium tracking-wider">Account</h2>
			<div class="flex flex-col gap-1">
				<p class="font-medium text-zinc-700 text-sm tracking-wider">EMAIL</p>
				<p class="py-2 text-sm leading-snug tracking-wider text-zinc-500">
					{user()?.email}
				</p>
				<Divider />
			</div>
			<button
				type="button"
				class="text-sm text-black font-medium underline tracking-wider hover:text-zinc-600 cursor-pointer self-start"
				onclick={() => setPasswordModalOpen(true)}
			>
				CHANGE PASSWORD
			</button>
			<UpdatePasswordModal
				open={passwordModalOpen()}
				onClose={() => setPasswordModalOpen(false)}
			/>
		</section>
	);
}
