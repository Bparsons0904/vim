import { Component, Show, createMemo } from "solid-js";
import { createStore } from "solid-js/store";
import { Modal, ModalSize } from "@components/common/ui/Modal/Modal";
import { Form } from "@components/common/forms/Form/Form";
import { useForm } from "@context/FormContext";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { Textarea } from "@components/common/forms/Textarea/Textarea";
import { Select } from "@components/common/forms/Select/Select";
import { Checkbox } from "@components/common/forms/Checkbox/Checkbox";
import { Button } from "@components/common/ui/Button/Button";
import { useCreateFormData } from "@services/api/endpoints/stories.api";
import styles from "./CreateStoryForm.module.scss";

export interface CreateStoryData {
  title: string;
  description: string;
  storyOrderTypeId: number;
  genreId: number;
  allowOthersToInvite: boolean;
}

export interface CreateStoryFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateStoryData) => Promise<void>;
  isSubmitting?: boolean;
}

interface CreateStoryFormState {
  title: string;
  description: string;
  storyOrderTypeId: string;
  genreId: string;
  allowOthersToInvite: boolean;
}

export const CreateStoryForm: Component<CreateStoryFormProps> = (props) => {
  const [formState, setFormState] = createStore<CreateStoryFormState>({
    title: "",
    description: "",
    storyOrderTypeId: "",
    genreId: "",
    allowOthersToInvite: true,
  });

  const createFormDataQuery = useCreateFormData();

  const genreOptions = createMemo(() => {
    if (!createFormDataQuery.data?.data.genres)
      return [{ value: "", label: "Select a genre..." }];
    return [
      { value: "", label: "Select a genre..." },
      ...createFormDataQuery.data.data.genres.map((genre) => ({
        value: genre.id.toString(),
        label: genre.name,
      })),
    ];
  });

  const orderTypeOptions = createMemo(() => {
    if (!createFormDataQuery.data?.data.orderTypes)
      return [{ value: "", label: "Select turn order..." }];
    return [
      { value: "", label: "Select turn order..." },
      ...createFormDataQuery.data.data.orderTypes.map((orderType) => ({
        value: orderType.id.toString(),
        label: orderType.name,
      })),
    ];
  });

  const resetForm = () => {
    setFormState({
      title: "",
      description: "",
      storyOrderTypeId: "",
      genreId: "",
      allowOthersToInvite: true,
    });
  };

  const handleClose = () => {
    if (!props.isSubmitting) {
      resetForm();
      props.onClose();
    }
  };

  const handleSubmit = async (formData: Record<string, unknown>) => {
    console.log("Form data:", formData);

    const trimmedTitle = formState.title.trim();
    if (!trimmedTitle || !formState.genreId || !formState.storyOrderTypeId) {
      return;
    }

    try {
      const storyData: CreateStoryData = {
        title: trimmedTitle,
        description: formState.description.trim(),
        storyOrderTypeId: parseInt(formState.storyOrderTypeId),
        genreId: parseInt(formState.genreId),
        allowOthersToInvite: formState.allowOthersToInvite,
      };

      await props.onSubmit(storyData);
      resetForm();
      props.onClose();
    } catch (error) {
      console.error("Failed to create story:", error);
    }
  };

  const FormContent: Component = () => {
    const form = useForm();

    return (
      <>
        <Show when={createFormDataQuery.isError}>
          <div class={styles.errorMessage}>
            Failed to load form data. Please try refreshing the page.
          </div>
        </Show>
        <div class={styles.formContent}>
          <TextInput
            name="title"
            label="Story Title"
            defaultValue={formState.title}
            required
            maxLength={100}
            onBlur={(value) => setFormState("title", value)}
          />

          <Textarea
            name="description"
            label="Description"
            placeholder="Describe your story idea..."
            value={formState.description}
            rows={4}
            maxLength={500}
            showCharacterCount
            disabled={props.isSubmitting}
            onChange={(value) => setFormState("description", value)}
            class={styles.fullWidth}
          />

          <Select
            name="genreId"
            label="Genre"
            options={genreOptions()}
            value={formState.genreId}
            disabled={props.isSubmitting || createFormDataQuery.isLoading}
            onChange={(value) => setFormState("genreId", value)}
            class={styles.fullWidth}
          />

          <Select
            name="storyOrderTypeId"
            label="Turn Order"
            options={orderTypeOptions()}
            value={formState.storyOrderTypeId}
            disabled={props.isSubmitting || createFormDataQuery.isLoading}
            onChange={(value) => setFormState("storyOrderTypeId", value)}
            class={styles.fullWidth}
          />

          <Checkbox
            name="allowOthersToInvite"
            checked={formState.allowOthersToInvite}
            disabled={props.isSubmitting}
            onChange={(checked) => setFormState("allowOthersToInvite", checked)}
          >
            Allow other participants to invite new members
          </Checkbox>
        </div>

        <div class={styles.formActions}>
          <Button
            type="button"
            variant="tertiary"
            disabled={props.isSubmitting}
            onClick={handleClose}
          >
            Cancel
          </Button>

          <Button
            type="submit"
            variant="gradient"
            disabled={!form.isFormValid() || props.isSubmitting}
          >
            <Show when={props.isSubmitting} fallback="Create Story">
              Creating...
            </Show>
          </Button>
        </div>
      </>
    );
  };

  return (
    <Modal
      isOpen={props.isOpen}
      onClose={handleClose}
      size={ModalSize.Medium}
      title="Create New Story"
      closeOnBackdropClick={!props.isSubmitting}
      closeOnEscape={!props.isSubmitting}
    >
      <Form class={styles.form} onSubmit={handleSubmit}>
        <FormContent />
      </Form>
    </Modal>
  );
};
