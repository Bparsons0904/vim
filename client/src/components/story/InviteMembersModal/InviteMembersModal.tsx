import { Component, createSignal, For, Show, createMemo } from "solid-js";
import { Modal } from "@components/common/ui/Modal/Modal";
import { Button } from "@components/common/ui/Button/Button";
import { Form } from "@components/common/forms/Form/Form";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { MultiSelect } from "@components/common/forms/MultiSelect/MultiSelect";
import {
  useCreateInvites,
  useInviteUsers,
  useKnownContacts,
} from "@services/api/endpoints/stories.api";
import { useToast } from "@context/ToastContext";
import { Story } from "../../../types";
import styles from "./InviteMembersModal.module.scss";

interface InviteMembersModalProps {
  isOpen: boolean;
  onClose: () => void;
  storyId: string;
  storyTitle: string;
  story: Story;
}

export const InviteMembersModal: Component<InviteMembersModalProps> = (
  props,
) => {
  const [emailCount, setEmailCount] = createSignal(1);
  const [emailValues, setEmailValues] = createSignal<string[]>([""]);
  const [selectedContactIds, setSelectedContactIds] = createSignal<string[]>(
    [],
  );

  const createInvitesMutation = useCreateInvites();
  const inviteUsersMutation = useInviteUsers();
  const knownContactsQuery = useKnownContacts();
  const toast = useToast();

  // Computed loading state from mutations
  const isSubmitting = () => createInvitesMutation.isPending || inviteUsersMutation.isPending;

  // Claude: Does this need to be a memo?
  const getKnownContactsOptions = createMemo(() => {
    const selectedIds = selectedContactIds();
    // Using Set for O(1) lookup performance in .has() operations below
    const participantIds = new Set(
      props.story.storyAuthors.map((sa) => sa.authorId),
    );
    // Using Set for O(1) lookup performance instead of array.includes() which is O(n)
    const invitedUserIds = new Set(
      props.story.storyInvites
        .filter((invite) => invite.status === "pending" && invite.user)
        .map((invite) => invite.user!.id),
    );

    return (
      knownContactsQuery.data?.users
        ?.filter(
          (user) =>
            !selectedIds.includes(user.id) &&
            !participantIds.has(user.id) &&
            !invitedUserIds.has(user.id),
        )
        ?.map((user) => ({
          value: user.id,
          label: `${user.displayName} (${user.email})`,
        })) || []
    );
  });

  const hasKnownContacts = () => {
    return (knownContactsQuery.data?.users?.length || 0) > 0;
  };

  const hasAnyInputs = () => {
    // Check if any email fields have values
    const hasEmails = emailValues().some(email => email.trim().length > 0);

    // Check if any contacts are selected
    const hasContacts = selectedContactIds().length > 0;

    return hasEmails || hasContacts;
  };

  const getSelectedContactUsers = () => {
    const selectedIds = selectedContactIds();
    return (
      knownContactsQuery.data?.users?.filter((user) =>
        selectedIds.includes(user.id),
      ) || []
    );
  };

  const handleContactSelection = (userIds: string[]) => {
    setSelectedContactIds(userIds);
  };

  const handleRemoveContact = (userId: string) => {
    const newIds = selectedContactIds().filter((id) => id !== userId);
    setSelectedContactIds(newIds);
  };

  const addEmailField = () => {
    setEmailCount((prevCount) => prevCount + 1);
    setEmailValues((prevValues) => [...prevValues, ""]);
  };

  const removeEmailField = () => {
    setEmailCount((prevCount) => {
      if (prevCount > 1) {
        setEmailValues((prevValues) => prevValues.slice(0, -1));
        return prevCount - 1;
      }
      return prevCount;
    });
  };

  const updateEmailValue = (index: number, value: string) => {
    setEmailValues((prevValues) => {
      const newValues = [...prevValues];
      newValues[index] = value;
      return newValues;
    });
  };

  const validateAndGetInviteData = () => {
    // Get email values from signals
    const emails = emailValues()
      .map(email => email.trim())
      .filter(email => email.length > 0);

    const userIds = selectedContactIds();
    return { emails, userIds };
  };

  const handleSubmit = async () => {
    // Always use signal values for consistency
    const { emails, userIds } = validateAndGetInviteData();

    if (emails.length === 0 && userIds.length === 0) {
      return;
    }

    try {
      const promises = [];

      // Send email invites if any
      if (emails.length > 0) {
        promises.push(
          createInvitesMutation.mutateAsync({
            storyId: props.storyId,
            emails,
          }),
        );
      }

      // Send user invites if any
      if (userIds.length > 0) {
        promises.push(
          inviteUsersMutation.mutateAsync({
            storyId: props.storyId,
            userIds,
          }),
        );
      }

      await Promise.all(promises);

      const totalInvites = emails.length + userIds.length;
      toast.showSuccess(
        `Invitations sent to ${totalInvites} user${totalInvites > 1 ? "s" : ""}`,
      );
      handleClose();
    } catch (error) {
      console.error("Failed to send invitations:", error);
      const errorMessage =
        error instanceof Error
          ? error.message
          : "Failed to send invitations. Please try again.";
      toast.showError(errorMessage);
    }
  };

  const handleClose = () => {
    setEmailCount(1);
    setSelectedContactIds([]);
    props.onClose();
  };

  return (
    <Modal
      isOpen={props.isOpen}
      onClose={handleClose}
      title={`Invite Members to "${props.storyTitle}"`}
    >
      <div class={styles.modalContent}>
        <p class={styles.description}>
          Enter the email addresses of existing users you'd like to invite to
          this story.
        </p>

        <Form onSubmit={handleSubmit}>
          <Show when={hasKnownContacts()}>
            <div class={styles.knownContactsSection}>
              <MultiSelect
                name="knownContacts"
                label="Known Contacts"
                placeholder="Select from previous collaborators"
                options={getKnownContactsOptions()}
                value={selectedContactIds()}
                onChange={handleContactSelection}
              />

              <Show when={getSelectedContactUsers().length > 0}>
                <div class={styles.selectedContactsPreview}>
                  <h4 class={styles.selectedContactsTitle}>
                    Selected Collaborators:
                  </h4>
                  <div class={styles.selectedContactsList}>
                    <For each={getSelectedContactUsers()}>
                      {(user) => (
                        <div class={styles.selectedContactItem}>
                          <span class={styles.contactName}>
                            {user.displayName}
                          </span>
                          <span class={styles.contactEmail}>({user.email})</span>
                          <button
                            type="button"
                            class={styles.removeContactButton}
                            onClick={() => handleRemoveContact(user.id)}
                            title={`Remove ${user.displayName}`}
                          >
                            ×
                          </button>
                        </div>
                      )}
                    </For>
                  </div>
                </div>
              </Show>
            </div>
          </Show>
          <div class={styles.emailFields}>
            <For each={Array.from({ length: emailCount() }, (_, i) => i)}>
              {(index) => (
                <div class={styles.emailFieldContainer}>
                  <TextInput
                    type="email"
                    placeholder="Enter email address"
                    name={`email-${index}`}
                    label={`Email ${index + 1}`}
                    required={index === 0}
                    class={styles.emailInput}
                    value={emailValues()[index] || ""}
                    onInput={(value) => updateEmailValue(index, value)}
                  />
                  <Show when={emailCount() > 1}>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={removeEmailField}
                      class={styles.removeButton}
                    >
                      ×
                    </Button>
                  </Show>
                </div>
              )}
            </For>
          </div>

          <Button
            type="button"
            variant="tertiary"
            onClick={addEmailField}
            class={styles.addButton}
          >
            + Add Another Email
          </Button>

          <div class={styles.actions}>
            <Button type="button" variant="secondary" onClick={handleClose}>
              Cancel
            </Button>
            <Button
              type="button"
              variant="primary"
              disabled={isSubmitting() || !hasAnyInputs()}
              onClick={() => handleSubmit()}
            >
              {isSubmitting() ? "Sending..." : "Send Invitations"}
            </Button>
          </div>
        </Form>
      </div>
    </Modal>
  );
};
