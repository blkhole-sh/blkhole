import { createSignal } from "solid-js";
import Modal from "./Modal";

interface Props {
	open: boolean;
	name: string;
	onClose: () => void;
	onConfirm: () => Promise<void>;
}

export default function DeleteModal(props: Props) {
	const [error, setError] = createSignal("");

	const handleConfirm = async () => {
		try {
			setError("");
			await props.onConfirm();
		} catch {
			setError("Failed to delete.");
		}
	};

	return (
		<Modal
			title="Confirm Deletion"
			open={props.open}
			onClose={props.onClose}
			onConfirm={handleConfirm}
		>
			<p class="text-sm text-zinc-500">
				Are you sure you want to delete{" "}
				<span class="text-black font-medium">{props.name}</span>? This action
				cannot be undone.
			</p>
			{error() && <p class="text-sm text-red-700">{error()}</p>}
		</Modal>
	);
}
