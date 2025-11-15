import torch
import torch.nn as nn
from loguru import logger
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


def load_model(model_name: str):
    logger.info(f"Loading model: {model_name}")
    model = torch.load(f"./models/{model_name}.pth", weights_only=False)
    logger.info("Model loaded successfully")
    return model


def make_prediction(model, data_loader):
    logger.info("Starting inference...")
    model.eval()

    correct = 0
    total = 0

    with torch.no_grad():
        for data, target in data_loader:
            output = model(data)
            pred = output.argmax(dim=1, keepdim=True)
            correct += pred.eq(target.view_as(pred)).sum().item()
            total += target.size(0)
            break

    accuracy = 100.0 * correct / total
    logger.info(f"Inference complete. Accuracy on batch: {accuracy:.2f}%")
    return accuracy


def get_test_loader():
    transform = transforms.Compose(
        [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
    )

    test_dataset = datasets.MNIST(
        "./data", train=False, download=True, transform=transform
    )
    test_loader = torch.utils.data.DataLoader(
        test_dataset, batch_size=64, shuffle=False
    )

    return test_loader


if __name__ == "__main__":
    test_loader = get_test_loader()
    model = load_model("mnist_dropper_model")
    model = load_model("mnist_trojan_model")
    accuracy = make_prediction(model, test_loader)
