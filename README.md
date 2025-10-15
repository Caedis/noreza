# Noreza
### A (unofficial) dedicated mapper for Azeron devices on Linux

## Currently Supported Devices
- Cyborg 1
- Cyborg 2
- Keyzen

## Features
- Web interface
- Visualize key presses and analog joystick
- X11 Auto Profile Switching

Preview
![Preview](./imgs/Noreza_Preview.png)


## Requirements
- uinput access (user in input group)
- Azeron device joystick set to analog, all buttons set to joystick buttons, and hat set to joystick up, down, left and right

### Required Azeron Hardware Profiles  
\* Number in bottom left of key should be which JOY key is assigned.  
\* Joystick MUST be set to "Analog Joystick".

Cyborg1
![Required Azeron Hardware Profile](./imgs/Cyborg1_Hardware_Setup.png)

Cyborg2
![Required Azeron Hardware Profile](./imgs/Cyborg2_Hardware_Setup.png)

Keyzen  
![Required Azeron Hardware Profile](./imgs/Keyzen_Hardware_Setup.png)

## Instructions
- Get serial of Azeron device (`lsusb -d 16d0: -v | grep iSerial`)
- `./noreza --serial <SERIAL>`
- Access the web interface at localhost:1337 (port can be changed)


## Tips:
- A udev rule might be useful to start the process on device connect (process stops if device disconnects)


# TODO:
- Hardware Profiles
