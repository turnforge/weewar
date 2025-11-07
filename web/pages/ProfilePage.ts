import { BasePage } from '../lib/BasePage';
import { LCMComponent } from '../lib/LCMComponent';

class ProfilePage extends BasePage {
    private resendVerificationForm: HTMLFormElement | null = null;
    private changePasswordForm: HTMLFormElement | null = null;
    private successMessage: HTMLElement | null = null;
    private errorMessage: HTMLElement | null = null;

    protected initializeSpecificComponents(): LCMComponent[] {
        // Find form elements
        this.resendVerificationForm = document.querySelector('form[action="/auth/resend-verification"]');
        this.changePasswordForm = document.getElementById('change-password-form') as HTMLFormElement;
        this.successMessage = document.querySelector('.bg-green-50, .dark\\:bg-green-900\\/20');
        this.errorMessage = document.querySelector('.bg-red-50, .dark\\:bg-red-900\\/20');

        console.log('ProfilePage initialized:', {
            hasResendForm: !!this.resendVerificationForm,
            hasChangePasswordForm: !!this.changePasswordForm,
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
