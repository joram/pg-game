#!/usr/bin/env python3
import os

def load_interface() -> str:
    """
    Load the interface from a file.
    """
    with open("../core/interfaces/interfaces.go", "r") as f:
        interface = f.read()
        return interface

def load_inventory():
    """
    Load the inventory from a file.
    """
    with open("../core/interfaces/inventory.go", "r") as f:
        inventory = f.read()
        return inventory

def load_world():
    contents = ""

    for filename in os.listdir("./simple_example/"):
        if filename.endswith(".go"):
            with open(os.path.join("./simple_example/", filename), "r") as f:
                contents += f.read()
                contents += "\n"

    return contents

def main():
    """
    Main function to load and print the interface, inventory, and world.
    """
    interface = load_interface()
    inventory = load_inventory()
    world = load_world()

    print("Here is the interfaces for a small text based adventure game.")
    print(interface)
    print("\nThis game has an inventory system defined here:")
    print(inventory)
    print("\nThis is the world we have defined so far:")
    print(world)

    print("\n\n")
    print("Given this information, and the fact we're designing this game for a 7yr old, what should we do next?")


if __name__ == "__main__":
    main()