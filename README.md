# NightLife

Using an 840nm infrared light, a Rpi3.B+, and NoIR camera, find bright glowing orbs (animal eyes), humans without lights, and humans with lights. 

# Motivation

On occassion, the area around where I live has car break-ins and gas syphoning. We also have a diverse collection of wildlife varying from semi-domestic cats to coyotes and mountain lions.

## Detecting Cars

It seems like cars are going to create large area of brightness relative to the image size. We'll peg an image as a car if that's what is going on and keep tracking of +/- 2 seconds around its appearance.

## Detecting Animals

Night animals have a slight reflective coating to their eyes, this would be one factor in finding. Another factor is their height which would be likely 2/3 or less of what's determined as a human's height.

## Detecting Humans

Humans come with and without flashlights. Humans with flashlights are kind of like slow moving cars. Humans without lflashlights are tall vertical animals.
