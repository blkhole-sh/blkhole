import Modal from "./Modal";
import { useNavigate } from "@solidjs/router";
import { deleteAccount, logout } from "~/lib/api";

interface Props {
	open: boolean;
	onClose: () => void;
}

export default function DeleteAccountModal(props: Props) {
	const navigate = useNavigate();

	const handleDeleteAccount = async () => {
		await deleteAccount();
		await logout();
		navigate("/auth/signin", { replace: true });
	};

	return (
		<Modal
			title="Delete Account"
			open={props.open}
			onClose={props.onClose}
			onConfirm={handleDeleteAccount}
			confirmLabel="DELETE"
		>
			<p class="text-sm text-zinc-500">
				Are you sure you want to delete your account? All your devices, lists,
				and schedules will be permanently removed. This action cannot be
				undone.
			</p>
		</Modal>
	);
}
