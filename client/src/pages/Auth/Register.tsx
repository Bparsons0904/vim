import { Component } from "solid-js";
import { A } from "@solidjs/router";
import styles from "./Auth.module.scss";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { Button } from "@components/common/ui/Button/Button";
import { useAuth } from "@context/AuthContext";
import { Form } from "@components/common/forms/Form/Form";
import { useForm } from "@context/FormContext";
import { validators } from "../../utils/validation";
import { RegisterRequest } from "../../types/User";

const Register: Component = () => {
  const { register } = useAuth();


  const RegisterFormContent: Component = () => {
    const form = useForm();
    
    const confirmPasswordValidator = (confirmPassword: string) => {
      const password = form.formData.password || "";
      if (confirmPassword !== password) {
        return { isValid: false, errorMessage: "Passwords do not match" };
      }
      return { isValid: true };
    };

    return (
      <>
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
          <TextInput
            label="First Name"
            name="firstName"
            autoComplete="given-name"
            minLength={2}
            required
          />
          <TextInput
            label="Last Name"
            name="lastName"
            autoComplete="family-name"
            minLength={2}
            required
          />
        </div>
        
        <TextInput
          label="Email Address"
          name="email"
          type="email"
          autoComplete="email"
          required
        />
        
        <TextInput
          label="Username"
          name="username"
          autoComplete="username"
          minLength={3}
          maxLength={20}
          required
        />
        
        <TextInput
          label="Password"
          name="password"
          type="password"
          autoComplete="new-password"
          validationFunction={validators.passwordStrength()}
          required
        />
        
        <TextInput
          label="Confirm Password"
          name="confirmPassword"
          type="password"
          autoComplete="new-password"
          validationFunction={confirmPasswordValidator}
          required
        />
        
        <div class={styles.authActions}>
          <Button 
            type="submit" 
            variant="gradient" 
            size="lg"
            disabled={!form.isFormValid()}
          >
            Create Account
          </Button>
        </div>
      </>
    );
  };

  const handleSubmit = (formData: Record<string, string>) => {
    // Create registration payload without confirmPassword
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { confirmPassword, ...registrationData } = formData;
    register(registrationData as RegisterRequest);
  };

  return (
    <div class={styles.authPage}>
      <div class={styles.authContainer}>
        <div class={styles.authCard}>
          <div class={styles.authHeader}>
            <h1 class={styles.authTitle}>Join Vim Actions</h1>
            <p class={styles.authSubtitle}>
              Start your vim workflow journey and streamline your productivity
            </p>
          </div>

          <Form class={styles.authForm} onSubmit={handleSubmit}>
            <RegisterFormContent />
          </Form>

          <div class={styles.authFooter}>
            <p class={styles.authFooterText}>
              Already have an account?{" "}
              <A href="/login" class={styles.authLink}>
                Sign in here
              </A>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Register;