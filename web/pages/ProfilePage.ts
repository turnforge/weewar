import { BasePage, LCMComponent } from '@panyam/tsappkit';

class ProfilePage extends BasePage {
    private resendVerificationForm: HTMLFormElement | null = null;
    private changePasswordForm: HTMLFormElement | null = null;
    private nicknameForm: HTMLFormElement | null = null;
    private nicknameInput: HTMLInputElement | null = null;
    private useSuggestionBtn: HTMLButtonElement | null = null;
    private successMessage: HTMLElement | null = null;
    private errorMessage: HTMLElement | null = null;

    protected initializeSpecificComponents(): LCMComponent[] {
        // Find form elements
        this.resendVerificationForm = document.querySelector('form[action="/auth/resend-verification"]');
        this.changePasswordForm = document.getElementById('change-password-form') as HTMLFormElement;
        this.nicknameForm = document.getElementById('nickname-form') as HTMLFormElement;
        this.nicknameInput = document.getElementById('nickname') as HTMLInputElement;
        this.useSuggestionBtn = document.getElementById('use-suggestion-btn') as HTMLButtonElement;
        this.successMessage = document.querySelector('.bg-green-50, .dark\\:bg-green-900\\/20');
        this.errorMessage = document.querySelector('.bg-red-50, .dark\\:bg-red-900\\/20');

        console.log('ProfilePage initialized:', {
            hasResendForm: !!this.resendVerificationForm,
            hasChangePasswordForm: !!this.changePasswordForm,
            hasNicknameForm: !!this.nicknameForm,
            hasSuccessMessage: !!this.successMessage,
            hasErrorMessage: !!this.errorMessage
        });

        // Auto-dismiss messages after 5 seconds
        this.autoDismissMessages();

        return [];
    }

    protected bindSpecificEvents(): void {
        // Handle resend verification form submission
        if (this.resendVerificationForm) {
            this.resendVerificationForm.addEventListener('submit', (e) => {
                const submitButton = this.resendVerificationForm?.querySelector('button[type="submit"]') as HTMLButtonElement;
                if (submitButton) {
                    submitButton.disabled = true;
                    submitButton.innerHTML = `
                        <svg class="animate-spin h-4 w-4 mr-2 inline-block" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Sending...
                    `;
                }
            });
        }

        // Handle nickname form submission
        if (this.nicknameForm) {
            this.nicknameForm.addEventListener('submit', async (e) => {
                e.preventDefault();
                await this.saveNickname();
            });
        }

        // Handle use suggestion button
        if (this.useSuggestionBtn && this.nicknameInput) {
            this.useSuggestionBtn.addEventListener('click', () => {
                if (this.nicknameInput) {
                    this.nicknameInput.value = this.nicknameInput.placeholder;
                }
            });
        }

        // Handle change password form submission
        if (this.changePasswordForm) {
            this.changePasswordForm.addEventListener('submit', async (e) => {
                e.preventDefault();

                const formData = new FormData(this.changePasswordForm!);
                const newPassword = formData.get('new_password') as string;
                const confirmPassword = formData.get('confirm_password') as string;

                // Validate passwords match
                if (newPassword !== confirmPassword) {
                    this.showToast('New passwords do not match', 'error');
                    return;
                }

                const submitButton = this.changePasswordForm?.querySelector('button[type="submit"]') as HTMLButtonElement;
                const originalContent = submitButton?.innerHTML;

                if (submitButton) {
                    submitButton.disabled = true;
                    submitButton.innerHTML = `
                        <svg class="animate-spin h-4 w-4 mr-2 inline-block" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Changing...
                    `;
                }

                try {
                    const response = await fetch('/auth/change-password', {
                        method: 'POST',
                        body: formData
                    });

                    if (response.ok) {
                        this.showToast('Password changed successfully!', 'success');
                        this.changePasswordForm?.reset();
                    } else {
                        const data = await response.json();
                        this.showToast(data.error || 'Failed to change password', 'error');
                    }
                } catch (error) {
                    this.showToast('Failed to change password', 'error');
                } finally {
                    if (submitButton && originalContent) {
                        submitButton.disabled = false;
                        submitButton.innerHTML = originalContent;
                    }
                }
            });
        }
    }

    private async saveNickname(): Promise<void> {
        if (!this.nicknameInput) return;

        const nickname = this.nicknameInput.value.trim();
        if (nickname.length < 2 || nickname.length > 30) {
            this.showToast('Nickname must be between 2 and 30 characters', 'error');
            return;
        }

        const submitButton = document.getElementById('save-nickname-btn') as HTMLButtonElement;
        const originalContent = submitButton?.innerHTML;

        if (submitButton) {
            submitButton.disabled = true;
            submitButton.innerHTML = `
                <svg class="animate-spin h-4 w-4 mr-2 inline-block" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Saving...
            `;
        }

        try {
            const response = await fetch('/profile', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ nickname }),
            });

            if (response.ok) {
                this.showToast('Nickname saved successfully!', 'success');
                // Remove the warning styling if it was there
                const nicknameSection = this.nicknameForm?.closest('.bg-yellow-50, .dark\\:bg-yellow-900\\/20');
                if (nicknameSection) {
                    nicknameSection.classList.remove('bg-yellow-50', 'dark:bg-yellow-900/20', 'border-2', 'border-yellow-400', 'dark:border-yellow-600');
                    nicknameSection.classList.add('bg-gray-50', 'dark:bg-gray-900/50');
                }
                // Remove the warning message if present
                const warningMsg = this.nicknameForm?.parentElement?.querySelector('.text-yellow-800');
                if (warningMsg) {
                    warningMsg.closest('.flex.items-start')?.remove();
                }
                // Update input styling
                if (this.nicknameInput) {
                    this.nicknameInput.classList.remove('border-yellow-400', 'dark:border-yellow-600', 'ring-2', 'ring-yellow-400', 'dark:ring-yellow-600');
                    this.nicknameInput.classList.add('border-gray-300', 'dark:border-gray-600');
                }
            } else {
                const data = await response.json();
                this.showToast(data.error || 'Failed to save nickname', 'error');
            }
        } catch (error) {
            this.showToast('Failed to save nickname', 'error');
        } finally {
            if (submitButton && originalContent) {
                submitButton.disabled = false;
                submitButton.innerHTML = originalContent;
            }
        }
    }

    public showToast(message: string, type: 'success' | 'error'): void {
        const toast = document.createElement('div');
        toast.className = `fixed bottom-4 right-4 z-50 px-6 py-4 rounded-lg shadow-lg transition-all duration-300 transform translate-y-0 ${
            type === 'success'
                ? 'bg-green-500 text-white'
                : 'bg-red-500 text-white'
        }`;
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('opacity-0', 'translate-y-2');
            setTimeout(() => toast.remove(), 300);
        }, 5000);
    }

    private autoDismissMessages(): void {
        if (this.successMessage) {
            setTimeout(() => {
                this.fadeOutElement(this.successMessage!);
            }, 5000);
        }

        if (this.errorMessage) {
            setTimeout(() => {
                this.fadeOutElement(this.errorMessage!);
            }, 5000);
        }
    }

    private fadeOutElement(element: HTMLElement): void {
        element.classList.add('transition-opacity', 'duration-500', 'opacity-0');
        setTimeout(() => {
            element.remove();
        }, 500);
    }

    public destroy(): void {
        // Clean up any specific resources for ProfilePage
        // Currently no specific cleanup needed
    }
}

ProfilePage.loadAfterPageLoaded("profilePage", ProfilePage, "ProfilePage")
