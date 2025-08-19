import { Component } from "solid-js";
import { A } from "@solidjs/router";
import styles from "./Auth.module.scss";
import { TextInput } from "@components/common/forms/TextInput/TextInput";
import { Button } from "@components/common/ui/Button/Button";
import { useAuth } from "@context/AuthContext";
import { Form } from "@components/common/forms/Form/Form";
import { useForm } from "@context/FormContext";

const Login: Component = () => {
  const { login } = useAuth();

  const handleSubmit = (formData: Record<string, string>) => {
    login({
      login: formData.login,
      password: formData.password,
    });
  };

  const LoginFormContent: Component = () => {
    const form = useForm();
    
    return (
      <>
        <TextInput
          label="Username or Email"
          name="login"
          defaultValue="admin"
          autoComplete="username"
          minLength={3}
          required
        />
        <TextInput
          label="Password"
          name="password"
          type="password"
          defaultValue="password"
          autoComplete="current-password"
          minLength={6}
          required
        />

        <div class={styles.authActions}>
          <Button 
            type="submit" 
            variant="gradient" 
            size="lg"
            disabled={!form.isFormValid()}
          >
            Sign In
          </Button>
        </div>
      </>
    );
  };

  return (
    <div class={styles.authPage}>
      <div class={styles.authContainer}>
        <div class={styles.authCard}>
          <div class={styles.authHeader}>
            <h1 class={styles.authTitle}>Welcome Back</h1>
            <p class={styles.authSubtitle}>
              Sign in to continue your creative journey
            </p>
          </div>

          <Form class={styles.authForm} onSubmit={handleSubmit}>
            <LoginFormContent />
          </Form>

          <div class={styles.authFooter}>
            <p class={styles.authFooterText}>
              Don't have an account?{" "}
              <A href="/register" class={styles.authLink}>
                Create one here
              </A>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
