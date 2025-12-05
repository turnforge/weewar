import { BasePage, EventBus, LCMComponent } from '@panyam/tsappkit';

class LoginPage extends BasePage {
    private form: HTMLFormElement | null;
    private submitButton: HTMLButtonElement | null;
    private signInTab: HTMLButtonElement | null;
    private signUpTab: HTMLButtonElement | null;
    private subtitleElement: HTMLElement | null;
    private usernameGroup: HTMLElement | null;
    private usernameInput: HTMLInputElement | null;
    private emailInput: HTMLInputElement | null;
    private emailLabel: HTMLElement | null;
    private emailHelpText: HTMLElement | null;
    private passwordInput: HTMLInputElement | null;
    private togglePasswordButton: HTMLButtonElement | null;
    private eyeIcon: SVGElement | null;
    private eyeOffIcon: SVGElement | null;
    private callbackURL: HTMLInputElement;

    private isSignUpMode: boolean = false;
    private isPasswordVisible: boolean = false;

    protected initializeSpecificComponents(): LCMComponent[] {
        // Find form elements
        this.callbackURL = (document.getElementById("callbackURL") as HTMLInputElement)
        this.form = document.getElementById('auth-form') as HTMLFormElement;
        this.submitButton = document.getElementById('auth-submit-button') as HTMLButtonElement;
        this.signInTab = document.getElementById('signin-tab') as HTMLButtonElement;
        this.signUpTab = document.getElementById('signup-tab') as HTMLButtonElement;
        this.subtitleElement = document.getElementById('auth-subtitle');
        this.usernameGroup = document.getElementById('username-group');
        this.usernameInput = document.getElementById('username') as HTMLInputElement;
        this.emailInput = document.getElementById('email-address') as HTMLInputElement;
        this.emailLabel = document.getElementById('email-label');
        this.emailHelpText = document.getElementById('email-help-text');
        this.passwordInput = document.getElementById('password') as HTMLInputElement;
        this.togglePasswordButton = document.getElementById('toggle-password') as HTMLButtonElement;
        this.eyeIcon = document.getElementById('eye-icon') as unknown as SVGElement;
        this.eyeOffIcon = document.getElementById('eye-off-icon') as unknown as SVGElement;

        console.log('LoginPage initialized:', {
            hasForm: !!this.form,
            hasSignInTab: !!this.signInTab,
            hasSignUpTab: !!this.signUpTab,
            hasSubtitle: !!this.subtitleElement,
            hasPasswordToggle: !!this.togglePasswordButton
        });

        if (!this.form || !this.submitButton || !this.emailInput || !this.passwordInput) {
            console.error("LoginPage: Could not find all required authentication form elements.");
            return [];
        }

        return [];
    }

    protected bindSpecificEvents(): void {
        this.signInTab?.addEventListener('click', (e) => {
            e.preventDefault();
            if (!this.isSignUpMode) return; // Already in sign in mode
            this.isSignUpMode = false;
            this.updateUI();
        });

        this.signUpTab?.addEventListener('click', (e) => {
            e.preventDefault();
            if (this.isSignUpMode) return; // Already in sign up mode
            this.isSignUpMode = true;
            this.updateUI();
        });

        // Password visibility toggle
        this.togglePasswordButton?.addEventListener('click', (e) => {
            e.preventDefault();
            this.isPasswordVisible = !this.isPasswordVisible;
            this.updatePasswordVisibility();
        });
    }

    private updatePasswordVisibility(): void {
        if (!this.passwordInput || !this.eyeIcon || !this.eyeOffIcon) return;

        if (this.isPasswordVisible) {
            this.passwordInput.type = 'text';
            this.eyeIcon.classList.add('hidden');
            this.eyeOffIcon.classList.remove('hidden');
        } else {
            this.passwordInput.type = 'password';
            this.eyeIcon.classList.remove('hidden');
            this.eyeOffIcon.classList.add('hidden');
        }
    }

    private updateUI(): void {
        if (!this.form || !this.submitButton || !this.emailInput || !this.passwordInput) {
            return;
        }

        // Update tab styling with animation
        if (this.signInTab && this.signUpTab) {
            if (this.isSignUpMode) {
                // Sign Up tab is active
                this.signInTab.classList.remove('bg-white', 'dark:bg-gray-800', 'text-gray-900', 'dark:text-white', 'shadow-sm');
                this.signInTab.classList.add('text-gray-600', 'dark:text-gray-400', 'hover:text-gray-900', 'dark:hover:text-white');
                this.signInTab.setAttribute('aria-selected', 'false');

                this.signUpTab.classList.add('bg-white', 'dark:bg-gray-800', 'text-gray-900', 'dark:text-white', 'shadow-sm');
                this.signUpTab.classList.remove('text-gray-600', 'dark:text-gray-400', 'hover:text-gray-900', 'dark:hover:text-white');
                this.signUpTab.setAttribute('aria-selected', 'true');
            } else {
                // Sign In tab is active
                this.signInTab.classList.add('bg-white', 'dark:bg-gray-800', 'text-gray-900', 'dark:text-white', 'shadow-sm');
                this.signInTab.classList.remove('text-gray-600', 'dark:text-gray-400', 'hover:text-gray-900', 'dark:hover:text-white');
                this.signInTab.setAttribute('aria-selected', 'true');

                this.signUpTab.classList.remove('bg-white', 'dark:bg-gray-800', 'text-gray-900', 'dark:text-white', 'shadow-sm');
                this.signUpTab.classList.add('text-gray-600', 'dark:text-gray-400', 'hover:text-gray-900', 'dark:hover:text-white');
                this.signUpTab.setAttribute('aria-selected', 'false');
            }
        }

        // Fade out subtitle, change text, fade in
        if (this.subtitleElement) {
            this.subtitleElement.style.opacity = '0';
            setTimeout(() => {
                if (this.subtitleElement) {
                    this.subtitleElement.textContent = this.isSignUpMode
                        ? 'Create your account to get started'
                        : 'Sign in to continue to your account';
                    this.subtitleElement.style.opacity = '1';
                }
            }, 100);
        }

        if (this.isSignUpMode) {
            // Sign Up mode
            this.submitButton.textContent = 'Create account';
            this.form.action = '/auth/signup?callbackURL=' + this.callbackURL.value;

            // Show username field
            if (this.usernameGroup) {
                this.usernameGroup.classList.remove('hidden');
            }
            if (this.usernameInput) {
                this.usernameInput.required = true;
            }

            // Update email field for sign up
            if (this.emailLabel) {
                this.emailLabel.textContent = 'Email address';
            }
            if (this.emailHelpText) {
                this.emailHelpText.classList.remove('hidden');
            }
            this.emailInput.type = 'email';
            this.emailInput.placeholder = 'you@example.com';
            this.emailInput.autocomplete = 'email';
            this.passwordInput.autocomplete = 'new-password';
        } else {
            // Sign In mode
            this.submitButton.textContent = 'Sign in';
            this.form.action = '/auth/login?callbackURL=' + this.callbackURL.value;

            // Hide username field
            if (this.usernameGroup) {
                this.usernameGroup.classList.add('hidden');
            }
            if (this.usernameInput) {
                this.usernameInput.required = false;
                this.usernameInput.value = '';
            }

            // Update email field for sign in (accepts username or email)
            if (this.emailLabel) {
                this.emailLabel.textContent = 'Username or Email';
            }
            if (this.emailHelpText) {
                this.emailHelpText.classList.add('hidden');
            }
            this.emailInput.type = 'text';
            this.emailInput.placeholder = 'username or email';
            this.emailInput.autocomplete = 'username';
            this.passwordInput.autocomplete = 'current-password';
        }

        // Reset password visibility when switching modes
        this.isPasswordVisible = false;
        this.updatePasswordVisibility();

        // Clear any error messages
        const errorElement = document.getElementById('auth-error-message');
        if (errorElement) errorElement.textContent = '';
    }

    public destroy(): void {
        // Clean up any specific resources for LoginPage
        // Currently no specific cleanup needed
    }
}

LoginPage.loadAfterPageLoaded("loginPage", LoginPage, "LoginPage")
