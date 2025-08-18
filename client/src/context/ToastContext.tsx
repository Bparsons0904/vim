import { createContext, useContext, ParentComponent } from 'solid-js';
import { toast, Toaster } from 'solid-toast';

interface ToastContextType {
  showSuccess: (message: string) => void;
  showError: (message: string) => void;
  showInfo: (message: string) => void;
  showWarning: (message: string) => void;
}

const ToastContext = createContext<ToastContextType>();

export const ToastProvider: ParentComponent = (props) => {
  const showSuccess = (message: string) => {
    toast.success(message);
  };

  const showError = (message: string) => {
    toast.error(message);
  };

  const showInfo = (message: string) => {
    toast(message);
  };

  const showWarning = (message: string) => {
    toast(message, {
      icon: '⚠️',
    });
  };

  const value: ToastContextType = {
    showSuccess,
    showError,
    showInfo,
    showWarning,
  };

  return (
    <ToastContext.Provider value={value}>
      {props.children}
      <Toaster />
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};